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
			ID:   "unknown",
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

	PermissionModuleMeta = []dto.Meta{
		{ID: "dashboard", Name: "Dashboard"},
		{ID: "employee", Name: "Pegawai"},
		{ID: "branch", Name: "Cabang"},
		{ID: "department", Name: "Departemen"},
		{ID: "position", Name: "Jabatan"},
		{ID: "role", Name: "Role"},
		{ID: "leave_type", Name: "Tipe Cuti"},
		{ID: "shift", Name: "Shift"},
		{ID: "holiday", Name: "Hari Libur"},
		{ID: "attendance", Name: "Kehadiran"},
		{ID: "leave", Name: "Cuti"},
		{ID: "request", Name: "Pengajuan"},
		{ID: "daily_report", Name: "RKH"},
		{ID: "mutabaah", Name: "Mutabaah"},
	}

	PermissionActionMeta = []dto.Meta{
		{ID: "read", Name: "Lihat"},
		{ID: "create", Name: "Tambah"},
		{ID: "update", Name: "Edit"},
		{ID: "delete", Name: "Hapus"},
		{ID: "approve", Name: "Approve"},
	}
)
