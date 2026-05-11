package service

import (
	"context"
	"fmt"
	"time"

	logger "hris-backend/config/log"
	"hris-backend/internal/repository"
	"hris-backend/internal/struct/dto"
	"hris-backend/internal/struct/model"
	"hris-backend/internal/utils/data"
)

// NotificationService — service untuk mengelola notifikasi & push
type NotificationService interface {
	// Subscription
	SaveSubscription(ctx context.Context, employeeID uint, req dto.PushSubscribeRequest) error
	GetSubscriptionStatus(ctx context.Context, employeeID uint) (bool, error)

	// Notification CRUD
	GetNotifications(ctx context.Context, employeeID uint, params dto.NotificationListParams) (dto.PaginatedResponse[dto.NotificationResponse], error)
	GetUnreadCount(ctx context.Context, employeeID uint) (int64, error)
	MarkAsRead(ctx context.Context, notificationID uint, employeeID uint) error
	MarkAllAsRead(ctx context.Context, employeeID uint) error

	// Business Triggers
	TriggerClockInReminder(ctx context.Context, employeeID uint, date, clockInEnd string) error
	TriggerClockOutReminder(ctx context.Context, employeeID uint, date, clockOutEnd string) error
	TriggerRequestApprovalNotification(ctx context.Context, requestType string, entityID uint, requesterEmployeeID uint) error
	TriggerApprovalResultNotification(ctx context.Context, requestType string, entityID uint, requesterEmployeeID uint, status string) error
	TriggerAbsentAlert(ctx context.Context, employeeID uint, date string) error
	TriggerMutabaahReminder(ctx context.Context, employeeID uint, date string) error

	// Cron Helpers
	SendMutabaahReminders(ctx context.Context, date string) error

	// DB-based push sender (dipanggil oleh cron tiap menit)
	SendPendingNotifications(ctx context.Context) error
}

type notificationService struct {
	pushRepo         repository.PushRepository
	notifRepo        repository.NotificationRepository
	pushSvc          PushService
	employeeRepo     repository.EmployeeRepository
	attendRepo       repository.AttendanceRepository
}

func NewNotificationService(
	pushRepo repository.PushRepository,
	notifRepo repository.NotificationRepository,
	pushSvc PushService,
	employeeRepo repository.EmployeeRepository,
	attendRepo repository.AttendanceRepository,
) NotificationService {
	return &notificationService{
		pushRepo:      pushRepo,
		notifRepo:     notifRepo,
		pushSvc:       pushSvc,
		employeeRepo:  employeeRepo,
		attendRepo:    attendRepo,
	}
}

// ============================================================================
// SUBSCRIPTION
// ============================================================================

func (s *notificationService) SaveSubscription(ctx context.Context, employeeID uint, req dto.PushSubscribeRequest) error {
	return s.pushRepo.SaveSubscription(ctx, nil, employeeID, req.Endpoint, req.Keys.P256dh, req.Keys.Auth, "")
}

func (s *notificationService) GetSubscriptionStatus(ctx context.Context, employeeID uint) (bool, error) {
	return s.pushRepo.IsEmployeeSubscribed(ctx, nil, employeeID)
}

// ============================================================================
// NOTIFICATION CRUD
// ============================================================================

func (s *notificationService) GetNotifications(ctx context.Context, employeeID uint, params dto.NotificationListParams) (dto.PaginatedResponse[dto.NotificationResponse], error) {
	return s.notifRepo.GetByEmployee(ctx, nil, employeeID, params)
}

func (s *notificationService) GetUnreadCount(ctx context.Context, employeeID uint) (int64, error) {
	return s.notifRepo.GetUnreadCount(ctx, nil, employeeID)
}

func (s *notificationService) MarkAsRead(ctx context.Context, notificationID uint, employeeID uint) error {
	return s.notifRepo.MarkAsRead(ctx, nil, notificationID, employeeID)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, employeeID uint) error {
	return s.notifRepo.MarkAllAsRead(ctx, nil, employeeID)
}

// ============================================================================
// WORDING BUILDER
// ============================================================================

func (s *notificationService) getRequestTypeName(requestType string) string {
	switch requestType {
	case "leave":
		return "Cuti"
	case "permission":
		return "Izin"
	case "overtime":
		return "Lembur"
	case "business_trip":
		return "Tugas"
	case "attendance_override":
		return "Koreksi Presensi"
	default:
		return "Pengajuan"
	}
}

func (s *notificationService) getRequestTypeDetail(requestType string, entityID uint) string {
	// This could be extended to fetch actual detail from DB
	return s.getRequestTypeName(requestType)
}

func (s *notificationService) buildApprovalTitle(requestType string) string {
	return fmt.Sprintf("Pengajuan %s Baru", s.getRequestTypeName(requestType))
}

func (s *notificationService) buildApprovalBody(requesterName, requestType string) string {
	return fmt.Sprintf("Ada pengajuan %s baru dari %s yang memerlukan persetujuan Anda.", s.getRequestTypeName(requestType), requesterName)
}

func (s *notificationService) buildResultTitle(requestType, status string) string {
	return fmt.Sprintf("Pengajuan %s %s", s.getRequestTypeName(requestType), s.formatStatus(status))
}

func (s *notificationService) buildResultBody(requestType, status string) string {
	if status == "approved" || status == "approved_leader" || status == "approved_hr" {
		return fmt.Sprintf("Selamat! Pengajuan %s Anda telah disetujui. Semoga harimu menyenangkan.", s.getRequestTypeName(requestType))
	}
	return fmt.Sprintf("Mohon maaf, pengajuan %s Anda belum dapat disetujui saat ini. Silakan hubungi atasan Anda untuk informasi lebih lanjut.", s.getRequestTypeName(requestType))
}

func (s *notificationService) formatStatus(status string) string {
	switch status {
	case "approved", "approved_leader", "approved_hr":
		return "Disetujui"
	case "rejected":
		return "Ditolak"
	default:
		return status
	}
}

// ============================================================================
// BUSINESS TRIGGERS
// ============================================================================

func (s *notificationService) TriggerClockInReminder(ctx context.Context, employeeID uint, date, clockInEnd string) error {
	title := "Pengingat Clock In"
	body := fmt.Sprintf("Selamat datang kembali! Jangan lupa untuk clock in sebelum jam %s ya.", clockInEnd)
	sendAt := s.calculateReminderSendAt(date, clockInEnd)
	return s.createNotification(ctx, employeeID, "clock_in_reminder", title, body, "/", "me", nil, nil, sendAt)
}

func (s *notificationService) TriggerClockOutReminder(ctx context.Context, employeeID uint, date, clockOutEnd string) error {
	title := "Pengingat Clock Out"
	body := fmt.Sprintf("Hari ini sudah produktif! Jangan lupa untuk clock out sebelum jam %s ya.", clockOutEnd)
	sendAt := s.calculateReminderSendAt(date, clockOutEnd)
	return s.createNotification(ctx, employeeID, "clock_out_reminder", title, body, "/", "me", nil, nil, sendAt)
}

// calculateReminderSendAt menghitung waktu kirim pengingat (10 menit sebelum batas waktu)
func (s *notificationService) calculateReminderSendAt(dateStr, timeStr string) time.Time {
	layout := "2006-01-02 15:04:05"
	datetimeStr := dateStr + " " + timeStr
	if len(timeStr) == 5 {
		datetimeStr = dateStr + " " + timeStr + ":00"
	}
	t, err := time.ParseInLocation(layout, datetimeStr, time.Local)
	if err != nil {
		// fallback ke sekarang jika parse gagal
		return time.Now()
	}
	return t.Add(-10 * time.Minute)
}

func (s *notificationService) TriggerRequestApprovalNotification(ctx context.Context, requestType string, entityID uint, requesterEmployeeID uint) error {
	// Get requester name
	emp, err := s.employeeRepo.GetEmployeeByID(ctx, nil, fmt.Sprintf("%d", requesterEmployeeID))
	if err != nil {
		return fmt.Errorf("get requester: %w", err)
	}

	recipients, err := s.notifRepo.GetRecipientsForApproval(ctx, nil, requesterEmployeeID)
	if err != nil {
		return fmt.Errorf("get recipients: %w", err)
	}

	title := s.buildApprovalTitle(requestType)
	body := s.buildApprovalBody(emp.FullName, requestType)

	for _, recipientID := range recipients {
		if err := s.createNotification(ctx, recipientID, requestType+"_request_new", title, body,
			data.NotificationActionURL[requestType], data.NotificationActionTab[requestType], &requestType, &entityID, time.Now()); err != nil {
			logger.Error("failed to create approval notification", map[string]any{
				"recipient_id": recipientID,
				"error":        err.Error(),
			})
		}
	}
	return nil
}

func (s *notificationService) TriggerApprovalResultNotification(ctx context.Context, requestType string, entityID uint, requesterEmployeeID uint, status string) error {
	title := s.buildResultTitle(requestType, status)
	body := s.buildResultBody(requestType, status)

	notifType := requestType + "_request_approved"
	if status == "rejected" {
		notifType = requestType + "_request_rejected"
	}

	return s.createNotification(ctx, requesterEmployeeID, notifType, title, body,
		data.NotificationActionURL[requestType], data.NotificationActionTab[requestType], &requestType, &entityID, time.Now())
}

func (s *notificationService) TriggerAbsentAlert(ctx context.Context, employeeID uint, date string) error {
	emp, err := s.employeeRepo.GetEmployeeByID(ctx, nil, fmt.Sprintf("%d", employeeID))
	if err != nil {
		return fmt.Errorf("get employee: %w", err)
	}

	recipients, err := s.notifRepo.GetRecipientsForApproval(ctx, nil, employeeID)
	if err != nil {
		return fmt.Errorf("get recipients: %w", err)
	}

	title := "Pegawai Tidak Hadir"
	body := fmt.Sprintf("%s tidak hadir tanpa keterangan hari ini. Mohon ditindaklanjuti ya.", emp.FullName)

	absentType := "absent_alert"
	for _, recipientID := range recipients {
		if err := s.createNotification(ctx, recipientID, absentType, title, body,
			"/attendance", "attendance_list", &absentType, &employeeID, time.Now()); err != nil {
			logger.Error("failed to create absent alert", map[string]any{
				"recipient_id": recipientID,
				"error":        err.Error(),
			})
		}
	}
	return nil
}

func (s *notificationService) TriggerMutabaahReminder(ctx context.Context, employeeID uint, date string) error {
	title := "Pengingat Mutabaah"
	body := "Jangan lupa untuk submit mutabaah hari ini ya."
	return s.createNotification(ctx, employeeID, "mutabaah_reminder", title, body, "/", "me", nil, nil, time.Now())
}

// ============================================================================
// HELPERS
// ============================================================================

func (s *notificationService) createNotification(
	ctx context.Context,
	employeeID uint,
	notifType, title, body, actionURL, actionTab string,
	relatedEntityType *string,
	relatedEntityID *uint,
	sendAt time.Time,
) error {
	n := &model.Notification{
		EmployeeID:        employeeID,
		Type:              notifType,
		Title:             title,
		Body:              body,
		ActionURL:         &actionURL,
		ActionTab:         &actionTab,
		PushStatus:        "pending",
		RelatedEntityType: relatedEntityType,
		RelatedEntityID:   relatedEntityID,
		SendAt:            sendAt,
	}
	if err := s.notifRepo.Create(ctx, nil, n); err != nil {
		return fmt.Errorf("create notification: %w", err)
	}

	// Notifikasi tersimpan di DB dengan status 'pending' dan send_at terjadwal.
	// Cron job akan polling dan mengirim push saat send_at <= NOW().
	return nil
}

// ============================================================================
// CRON HELPERS
// ============================================================================

// SendPendingNotifications — dipanggil oleh cron setiap menit.
// Mengambil notifikasi pending dari DB, kirim push, dan update status.
func (s *notificationService) SendPendingNotifications(ctx context.Context) error {
	notifications, err := s.notifRepo.GetPendingNotifications(ctx, nil, 50)
	if err != nil {
		return fmt.Errorf("get pending notifications: %w", err)
	}

	if len(notifications) == 0 {
		return nil
	}

	logger.Info("cron: sending pending notifications", map[string]any{
		"count": len(notifications),
	})

	for _, notif := range notifications {
		if err := s.sendSingleNotification(ctx, &notif); err != nil {
			logger.Error("failed to send notification", map[string]any{
				"notification_id": notif.ID,
				"error":           err.Error(),
			})
		}
	}

	return nil
}

func (s *notificationService) sendSingleNotification(ctx context.Context, notif *model.Notification) error {
	// Conditional check for clock reminders
	if s.shouldSkipClockReminder(ctx, notif) {
		_ = s.notifRepo.UpdatePushStatus(ctx, nil, notif.ID, "skipped")
		return nil
	}

	// Get active subscriptions
	subs, err := s.pushRepo.GetActiveSubscriptionsByEmployeeID(ctx, nil, notif.EmployeeID)
	if err != nil {
		_ = s.incrementAttemptAndMaybeFail(ctx, notif.ID)
		return fmt.Errorf("get subscriptions: %w", err)
	}

	if len(subs) == 0 {
		_ = s.notifRepo.UpdatePushStatus(ctx, nil, notif.ID, "no_subscription")
		return nil
	}

	// Build payload
	actionURL := ""
	if notif.ActionURL != nil {
		actionURL = *notif.ActionURL
	}
	actionTab := ""
	if notif.ActionTab != nil {
		actionTab = *notif.ActionTab
	}
	payload := BuildPushPayload(notif.Title, notif.Body, actionURL, actionTab, notif.ID)

	// Send to all subscriptions
	anySuccess := false
	for _, sub := range subs {
		if err := s.pushSvc.SendPush(ctx, sub, payload); err != nil {
			logger.Warn("push send failed", map[string]any{
				"notification_id": notif.ID,
				"error":           err.Error(),
			})
			// Deactivate expired subscriptions
			if err.Error() != "" && (containsStr(err.Error(), "expired") || containsStr(err.Error(), "410")) {
				_ = s.pushRepo.DeactivateSubscription(ctx, nil, sub.Endpoint)
			}
		} else {
			anySuccess = true
		}
	}

	if anySuccess {
		return s.notifRepo.UpdatePushStatus(ctx, nil, notif.ID, "sent")
	}

	return s.incrementAttemptAndMaybeFail(ctx, notif.ID)
}

func (s *notificationService) incrementAttemptAndMaybeFail(ctx context.Context, notifID uint) error {
	if err := s.notifRepo.IncrementPushAttempts(ctx, nil, notifID); err != nil {
		return err
	}

	notif, err := s.notifRepo.GetByID(ctx, nil, notifID)
	if err != nil {
		return err
	}

	if notif.PushAttempts >= 3 {
		_ = s.notifRepo.UpdatePushStatus(ctx, nil, notifID, "failed")
		logger.Warn("cron: max attempts reached", map[string]any{
			"notification_id": notifID,
		})
	}
	return nil
}

func (s *notificationService) shouldSkipClockReminder(ctx context.Context, notif *model.Notification) bool {
	// Clock in reminder: skip if already clocked in
	if notif.Type == "clock_in_reminder" {
		log, _ := s.attendRepo.GetTodayLog(ctx, nil, notif.EmployeeID, notif.CreatedAt.Format("2006-01-02"))
		if log != nil && (log.Status == string(model.AttendancePresent) || log.Status == string(model.AttendanceLate) || log.Status == string(model.AttendanceHalfDay)) {
			return true
		}
		return false
	}
	// Clock out reminder: skip if not clocked in or already clocked out
	if notif.Type == "clock_out_reminder" {
		log, _ := s.attendRepo.GetTodayLog(ctx, nil, notif.EmployeeID, notif.CreatedAt.Format("2006-01-02"))
		if log == nil {
			return true // not clocked in yet
		}
		if log.ClockOutAt != nil {
			return true // already clocked out
		}
		if log.Status == string(model.AttendanceAbsent) || log.Status == string(model.AttendanceLeave) {
			return true
		}
		return false
	}
	return false
}

func containsStr(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstr(s, substr)))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// SendMutabaahReminders — called by cron at 12:00 and 18:00
func (s *notificationService) SendMutabaahReminders(ctx context.Context, date string) error {
	employees, err := s.notifRepo.GetEmployeesForMutabaahReminder(ctx, nil, date)
	if err != nil {
		return fmt.Errorf("get mutabaah reminder employees: %w", err)
	}
	for _, empID := range employees {
		if err := s.TriggerMutabaahReminder(ctx, empID, date); err != nil {
			logger.Error("failed to trigger mutabaah reminder", map[string]any{
				"employee_id": empID,
				"error":       err.Error(),
			})
		}
	}
	return nil
}
