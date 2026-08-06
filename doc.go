// Package dexcomshare provides a client for the Dexcom Share API, which serves
// near-realtime glucose readings uploaded from a Dexcom continuous glucose
// monitor.
//
// The Share API is undocumented and unsupported by Dexcom. It can change or
// disappear without notice. You have been warned.
//
// Accounts registered outside the United States live on a separate host; pass
// [WithBaseURL] with [BaseURLOutsideUS] for those.
package dexcomshare
