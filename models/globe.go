package models

var (
	HOST_NATIVE = "native"

	HEALTH_DEGRADED = "degraded"
	HEALTH_DOWN     = "down"
	HEALTH_FAILED   = "failed"
	HEALTH_NORMAL   = "normal"
	HEALTH_PARTIAL  = "partial"

	ROLE_DATA         = "data"
	ROLE_DATA_SPARE   = "data&spare"
	ROLE_GLOBAL_SPARE = "global_spare"
	ROLE_RESERVED     = "reserved"
	ROLE_SPARE        = "spare"
	ROLE_UNUSED       = "unused"
	ROLE_KICKED       = "kicked"

	DISKTYPE_SATA    = "sata"
	DISKTYPE_SAS     = "sas"
	DISKTYPE_SSD     = "ssd"
	LEVEL            = map[int64]bool{0: true, 1: true, 5: true, 6: true}
	REBUILD_PRIORITY = map[string]bool{"low": true, "medium": true, "high": true}
)
