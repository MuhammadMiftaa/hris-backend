package data

import "time"

var (
	DEVELOPMENT_MODE = "development"
	STAGING_MODE     = "staging"
	PRODUCTION_MODE  = "production"

	SESSION_TTL = 3 * 24 * time.Hour

	// REQUEST_ID_HEADER is the standard header name used to propagate request IDs.
	REQUEST_ID_HEADER = "X-Request-ID"
	// REQUEST_ID_LOCAL_KEY is the key used to store the request ID in Gin's context locals.
	REQUEST_ID_LOCAL_KEY = "request_id"
)
