package repository

import (
	"context"
	"errors"
	"time"

	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"
	"hris-backend/internal/utils"

	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(ctx context.Context, tx Transaction, n *model.Notification) error
	GetByID(ctx context.Context, tx Transaction, id uint) (*model.Notification, error)
	GetByEmployee(ctx context.Context, tx Transaction, employeeID uint, params dto.NotificationListParams) (dto.PaginatedResponse[dto.NotificationResponse], error)
	GetUnreadCount(ctx context.Context, tx Transaction, employeeID uint) (int64, error)
	MarkAsRead(ctx context.Context, tx Transaction, notificationID uint, employeeID uint) error
	MarkAllAsRead(ctx context.Context, tx Transaction, employeeID uint) error
	UpdatePushStatus(ctx context.Context, tx Transaction, id uint, status string) error
	IncrementPushAttempts(ctx context.Context, tx Transaction, id uint) error
	GetRecipientsForApproval(ctx context.Context, tx Transaction, requesterEmployeeID uint) ([]uint, error)
	GetEmployeesForMutabaahReminder(ctx context.Context, tx Transaction, date string) ([]uint, error)
	CheckEmployeePermission(ctx context.Context, tx Transaction, employeeID uint, permCode string) (bool, error)
	// DB-based polling for push sender (no Redis)
	GetPendingNotifications(ctx context.Context, tx Transaction, limit int) ([]model.Notification, error)
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) getDB(ctx context.Context, tx Transaction) (*gorm.DB, error) {
	if tx != nil {
		gormTx, ok := tx.(*GormTx)
		if !ok {
			return nil, errors.New("invalid transaction type")
		}
		return gormTx.db.WithContext(ctx), nil
	}
	return r.db.WithContext(ctx), nil
}

func (r *notificationRepository) Create(ctx context.Context, tx Transaction, n *model.Notification) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Create(n).Error
}

func (r *notificationRepository) GetByID(ctx context.Context, tx Transaction, id uint) (*model.Notification, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}
	var n model.Notification
	if err := db.Where("id = ? AND deleted_at IS NULL", id).First(&n).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *notificationRepository) GetByEmployee(ctx context.Context, tx Transaction, employeeID uint, params dto.NotificationListParams) (dto.PaginatedResponse[dto.NotificationResponse], error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return dto.PaginatedResponse[dto.NotificationResponse]{}, err
	}

	baseQuery := `FROM notifications WHERE employee_id = ? AND deleted_at IS NULL AND send_at <= ?`
	args := []interface{}{employeeID, utils.NowWIB()}

	if params.IsRead != nil {
		baseQuery += " AND is_read = ?"
		args = append(args, *params.IsRead)
	}

	var total int
	if err := db.Raw("SELECT COUNT(*) "+baseQuery, args...).Scan(&total).Error; err != nil {
		return dto.PaginatedResponse[dto.NotificationResponse]{}, err
	}

	selectQuery := `SELECT
		id, type, title, body, action_url, action_tab,
		is_read, read_at, push_status,
		related_entity_type, related_entity_id,
		send_at
	` + baseQuery

	selectQuery += utils.BuildSortClause("notification", params.SortBy, params.GetSortDir(), "send_at DESC")
	selectQuery += utils.BuildPaginationClause(params.PaginationParams)

	var rows []dto.NotificationResponse
	if err := db.Raw(selectQuery, args...).Scan(&rows).Error; err != nil {
		return dto.PaginatedResponse[dto.NotificationResponse]{}, err
	}

	if rows == nil {
		rows = []dto.NotificationResponse{}
	}

	perPage := params.GetPerPage()
	totalPage := 1
	if perPage > 0 && total > 0 {
		totalPage = (total + perPage - 1) / perPage
	}

	return dto.PaginatedResponse[dto.NotificationResponse]{
		Data: rows,
		Pagination: dto.PaginationMeta{
			Page:      params.GetPage(),
			PerPage:   perPage,
			Total:     total,
			TotalPage: totalPage,
		},
	}, nil
}

func (r *notificationRepository) GetUnreadCount(ctx context.Context, tx Transaction, employeeID uint) (int64, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return 0, err
	}
	var count int64
	err = db.Model(&model.Notification{}).
		Where("employee_id = ? AND is_read = FALSE AND deleted_at IS NULL AND send_at <= ?", employeeID, utils.NowWIB()).
		Count(&count).Error
	return count, err
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, tx Transaction, notificationID uint, employeeID uint) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Model(&model.Notification{}).
		Where("id = ? AND employee_id = ? AND deleted_at IS NULL", notificationID, employeeID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": utils.NowWIB(),
		}).Error
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, tx Transaction, employeeID uint) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Model(&model.Notification{}).
		Where("employee_id = ? AND is_read = FALSE AND deleted_at IS NULL", employeeID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": utils.NowWIB(),
		}).Error
}

func (r *notificationRepository) UpdatePushStatus(ctx context.Context, tx Transaction, id uint, status string) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Model(&model.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"push_status":     status,
			"last_attempt_at": utils.NowWIB(),
		}).Error
}

func (r *notificationRepository) IncrementPushAttempts(ctx context.Context, tx Transaction, id uint) error {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return err
	}
	return db.Model(&model.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"push_attempts":   gorm.Expr("push_attempts + 1"),
			"last_attempt_at": utils.NowWIB(),
		}).Error
}

func (r *notificationRepository) GetRecipientsForApproval(ctx context.Context, tx Transaction, requesterEmployeeID uint) ([]uint, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	// Ambil department_id requester
	var reqDeptID *uint
	if err := db.Raw(`SELECT department_id FROM employees WHERE id = ? AND deleted_at IS NULL`, requesterEmployeeID).Scan(&reqDeptID).Error; err != nil {
		return nil, err
	}

	var ids []uint
	query := `
		SELECT DISTINCT e.id
		FROM employees e
		JOIN accounts a ON a.employee_id = e.id AND a.deleted_at IS NULL AND a.is_active = TRUE
		JOIN roles r ON r.id = a.role_id AND r.deleted_at IS NULL
		WHERE e.deleted_at IS NULL
		  AND (
			  r.level IN ('superadmin', 'admin')
			  OR (r.level = 'manager' AND e.department_id = ?)
		  )
		  AND e.id != ?
	`
	args := []interface{}{}
	if reqDeptID != nil {
		args = append(args, *reqDeptID)
	} else {
		args = append(args, nil)
	}
	args = append(args, requesterEmployeeID)

	if err := db.Raw(query, args...).Scan(&ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *notificationRepository) GetEmployeesForMutabaahReminder(ctx context.Context, tx Transaction, date string) ([]uint, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	var ids []uint
	// Pegawai yang hadir (present/late) tapi belum submit mutabaah di tanggal tersebut
	// Hanya pegawai dengan permission PERM_MutabaahCreate
	err = db.Raw(`
		SELECT DISTINCT al.employee_id
		FROM attendance_logs al
		LEFT JOIN mutabaah_logs ml
			ON ml.employee_id = al.employee_id
			AND ml.log_date = al.attendance_date
			AND ml.deleted_at IS NULL
		JOIN accounts a ON a.employee_id = al.employee_id AND a.deleted_at IS NULL AND a.is_active = TRUE
		JOIN role_permissions rp ON rp.role_id = a.role_id AND rp.permission_code = 'mutabaah-create'
		WHERE al.attendance_date = ?::DATE
		  AND al.status IN ('present', 'late')
		  AND al.deleted_at IS NULL
		  AND (ml.id IS NULL OR ml.is_submitted = FALSE)
	`, date).Scan(&ids).Error
	return ids, err
}

func (r *notificationRepository) CheckEmployeePermission(ctx context.Context, tx Transaction, employeeID uint, permCode string) (bool, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return false, err
	}
	var count int64
	err = db.Raw(`
		SELECT COUNT(*)
		FROM accounts a
		JOIN role_permissions rp ON rp.role_id = a.role_id
		WHERE a.employee_id = ? AND a.deleted_at IS NULL AND a.is_active = TRUE
		  AND rp.permission_code = ?
	`, employeeID, permCode).Scan(&count).Error
	return count > 0, err
}

// GetPendingNotifications — ambil notifikasi yang perlu dikirim via push.
// Criteria:
//   - push_status = 'pending' dan belum pernah dicoba (attempts = 0)
//   - ATAU push_status = 'failed', attempts < 3, dan sudah lewat 2 menit sejak last_attempt_at
// Dibatasi limit untuk menghindari overload dalam satu tick.
func (r *notificationRepository) GetPendingNotifications(ctx context.Context, tx Transaction, limit int) ([]model.Notification, error) {
	db, err := r.getDB(ctx, tx)
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 50
	}

	var notifications []model.Notification
	err = db.Where(`
		deleted_at IS NULL
		AND push_status IN ('pending', 'failed')
		AND push_attempts < 3
		AND send_at <= ?
		AND (
			last_attempt_at IS NULL
			OR last_attempt_at < ?
		)
	`, utils.NowWIB(), utils.NowWIB().Add(-2*time.Minute)).Order("send_at ASC, created_at ASC").Limit(limit).Find(&notifications).Error
	return notifications, err
}
