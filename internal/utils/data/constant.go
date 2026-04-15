package data

import "time"

var (
	DEVELOPMENT_MODE = "development"
	STAGING_MODE     = "staging"
	PRODUCTION_MODE  = "production"

	DateFormats = []string{
		"2006-01-02",
		"02/01/2006",
		"01/02/2006",
		"02-01-2006",
		"January 2, 2006",
		"Jan 2, 2006",
	}

	TimestampFormats = []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05.999999",
		"02/01/2006 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}

	SESSION_TTL = 3 * 24 * time.Hour

	// REQUEST_ID_HEADER is the standard header name used to propagate request IDs.
	REQUEST_ID_HEADER = "X-Request-ID"
	// REQUEST_ID_LOCAL_KEY is the key used to store the request ID in Gin's context locals.
	REQUEST_ID_LOCAL_KEY = "request_id"
)
