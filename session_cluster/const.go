package session

import "time"

const defaultSessionKeyName = "y"
const defaultDomain = ""
const defaultExpires = time.Duration(15) * time.Minute
const renewExpires = time.Duration(30) * time.Minute
const defaultGCLifetime = time.Duration(60) * time.Second
const defaultSecure = false
const defaultSessionIDInURLQuery = false
const defaultSessionIDInHTTPHeader = false

const expirationAttributeKey = "_sid_"
