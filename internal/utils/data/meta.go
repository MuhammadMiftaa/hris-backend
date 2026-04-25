package data

import "hris-backend/internal/struct/dto"

var (
	GenderMeta = []dto.Meta{
		{
			ID:   "male",
			Name: "Laki-laki",
		},
		{
			ID:   "female",
			Name: "Perempuan",
		},
	}

	MaritalStatusMeta = []dto.Meta{
		{
			ID:   "single",
			Name: "Belum Menikah",
		},
		{
			ID:   "married",
			Name: "Menikah",
		},
		{
			ID:   "divorced",
			Name: "Bercerai",
		},
		{
			ID:   "widowed",
			Name: "Duda/Janda",
		},
	}

	ReligionMeta = []dto.Meta{
		{
			ID:   "islam",
			Name: "Islam",
		},
		{
			ID:   "kristen",
			Name: "Kristen",
		},
		{
			ID:   "katolik",
			Name: "Katolik",
		},
		{
			ID:   "hindu",
			Name: "Hindu",
		},
		{
			ID:   "budha",
			Name: "Budha",
		},
		{
			ID:   "lainnya",
			Name: "Lainnya",
		},
	}

	BloodTypeMeta = []dto.Meta{
		{
			ID:   "a",
			Name: "A",
		},
		{
			ID:   "b",
			Name: "B",
		},
		{
			ID:   "ab",
			Name: "AB",
		},
		{
			ID:   "o",
			Name: "O",
		},
		{
			ID:   "-",
			Name: "Tidak Diketahui",
		},
	}

	StatusMeta = []dto.Meta{
		{
			ID:   "active",
			Name: "Aktif",
		},
		{
			ID:   "inactive",
			Name: "Tidak Aktif",
		},
	}

	LeaveCategoryMeta = []dto.Meta{
		{ID: "annual", Name: "Cuti Tahunan"},
		{ID: "sick", Name: "Sakit"},
		{ID: "maternity", Name: "Cuti Melahirkan"},
		{ID: "paternity", Name: "Cuti Ayah"},
		{ID: "unpaid", Name: "Tanpa Gaji"},
		{ID: "other", Name: "Lainnya"},
	}

	DurationUnitMeta = []dto.Meta{
		{ID: "days", Name: "Hari"},
		{ID: "hours", Name: "Jam"},
	}

	HolidayTypeMeta = []dto.Meta{
		{ID: "national", Name: "Nasional"},
		{ID: "joint", Name: "Cuti Bersama"},
		{ID: "observance", Name: "Peringatan"},
		{ID: "company", Name: "Perusahaan"},
	}

	ContractTypeMeta = []dto.Meta{
		{ID: "pkwt", Name: "PKWT"},
		{ID: "pkwtt", Name: "PKWTT"},
		{ID: "probation", Name: "Probation"},
		{ID: "intern", Name: "Intern"},
		{ID: "part_time", Name: "Part Time"},
		{ID: "freelance", Name: "Freelance"},
	}

	ContactTypeMeta = []dto.Meta{
		{ID: "phone", Name: "Telepon"},
		{ID: "email", Name: "Email"},
		{ID: "address", Name: "Alamat"},
	}

	DayOfWeekMeta = []dto.Meta{
		{ID: "monday", Name: "Senin"},
		{ID: "tuesday", Name: "Selasa"},
		{ID: "wednesday", Name: "Rabu"},
		{ID: "thursday", Name: "Kamis"},
		{ID: "friday", Name: "Jum'at"},
		{ID: "saturday", Name: "Sabtu"},
		{ID: "sunday", Name: "Minggu"},
	}

	RoleLevelMeta = []dto.Meta{
		{ID: "superadmin", Name: "Super Admin"},
		{ID: "admin", Name: "Admin"},
		{ID: "manager", Name: "Manager"},
		{ID: "staff", Name: "Staff"},
	}

	PermissionModuleMeta = []dto.Meta{
		{ID: "home", Name: "Home"},
		{ID: "employee", Name: "Pegawai"},
		{ID: "branch", Name: "Cabang"},
		{ID: "department", Name: "Departemen"},
		{ID: "job_position", Name: "Jabatan"},
		{ID: "role", Name: "Role"},
		{ID: "leave_type", Name: "Jenis Cuti"},
		{ID: "template_shift", Name: "Template Shift"},
		{ID: "schedule", Name: "Jadwal Pegawai"},
		{ID: "holiday", Name: "Hari Libur"},
		{ID: "attendance", Name: "Presensi"},
		{ID: "attendance_adjustment", Name: "Koreksi Presensi"},
		{ID: "leave", Name: "Cuti"},
		{ID: "leave_balance", Name: "Saldo Cuti"},
		{ID: "request", Name: "Pengajuan"},
		{ID: "business_trip", Name: "Tugas"},
		{ID: "overtime", Name: "Lembur"},
		{ID: "daily_report", Name: "RKH"},
		{ID: "mutabaah", Name: "Mutabaah"},
		{ID: "profile", Name: "Profile"},
	}

	PermissionActionMeta = []dto.Meta{
		{ID: "read", Name: "Lihat"},
		{ID: "employee_read", Name: "Lihat Khusus Pegawai"},
		{ID: "admin_read", Name: "Lihat Khusus Admin"},
		{ID: "create", Name: "Tambah"},
		{ID: "update", Name: "Edit"},
		{ID: "delete", Name: "Hapus"},
		{ID: "approve", Name: "Approve"},
		{ID: "leader_approve", Name: "Approve Leader"},
		{ID: "hr_approve", Name: "Approve HR"},
		{ID: "clock_in", Name: "Clock In"},
		{ID: "clock_out", Name: "Clock Out"},
		{ID: "reset_password", Name: "Reset Password"},
	}

	AttendanceStatusMeta = []dto.Meta{
		{ID: "present", Name: "Hadir"},
		{ID: "late", Name: "Terlambat"},
		{ID: "absent", Name: "Absen"},
		{ID: "half_day", Name: "Setengah Hari"},
		{ID: "leave", Name: "Cuti"},
		{ID: "business_trip", Name: "Dinas Luar"},
		{ID: "holiday", Name: "Libur"},
	}

	ClockMethodMeta = []dto.Meta{
		{ID: "gps", Name: "GPS"},
		{ID: "qr_code", Name: "QR Code"},
		{ID: "face_recognition", Name: "Face Recognition"},
		{ID: "manual", Name: "Manual"},
	}

	OverrideTypeMeta = []dto.Meta{
		{ID: "clock_in", Name: "Koreksi Jam Masuk"},
		{ID: "clock_out", Name: "Koreksi Jam Keluar"},
		{ID: "full_day", Name: "Koreksi Penuh"},
	}

	OverrideStatusMeta = []dto.Meta{
		{ID: "pending", Name: "Menunggu"},
		{ID: "approved", Name: "Disetujui"},
		{ID: "rejected", Name: "Ditolak"},
	}

	PermissionTypeMeta = []dto.Meta{
		{ID: "late_arrival", Name: "Izin Terlambat"},
		{ID: "early_leave", Name: "Izin Pulang Cepat"},
		{ID: "out_of_office", Name: "Izin Keluar Kantor"},
	}

	WorkLocationMeta = []dto.Meta{
		{ID: "office", Name: "Kantor"},
		{ID: "home", Name: "Rumah"},
		{ID: "outside", Name: "Luar Kantor"},
	}

	OvertimeStatusMeta = []dto.Meta{
		{ID: "pending", Name: "Menunggu"},
		{ID: "approved_leader", Name: "Disetujui Leader"},
		{ID: "approved_hr", Name: "Disetujui HR"},
		{ID: "rejected", Name: "Ditolak"},
	}

	TripStatusMeta = []dto.Meta{
		{ID: "pending", Name: "Menunggu"},
		{ID: "approved", Name: "Disetujui"},
		{ID: "rejected", Name: "Ditolak"},
	}

	LeaveRequestStatusMeta = []dto.Meta{
		{ID: "pending", Name: "Menunggu"},
		{ID: "approved_leader", Name: "Disetujui Leader"},
		{ID: "approved_hr", Name: "Disetujui HR"},
		{ID: "rejected", Name: "Ditolak"},
	}
)
