package entities

// XForwardedForKey use for get X-Forwarded-For header
// it is change to variable because in some case like performance test or stress test
// we should change it to Fake-X-Forwarded-For
var XForwardedForKey string = "X-Forwarded-For"
