package net

import (
	"os/exec"
	"runtime"
)

// DNSOSCache represents DNS caching system that Operating System configures.
type DNSOSCache struct {
	logger Logger
	logTag string
}

func NewDNSOSCache(logger Logger) DNSOSCache {
	return DNSOSCache{logger, "dns.DNSOSCache"}
}

func (c DNSOSCache) Flush() {
	switch runtime.GOOS {
	case "darwin":
		// Most OS Xs have mDNSResponder which caches entries going thru native DNS resolution.
		// If cache isnt cleared before our own DNS resolution takes over, following case may
		// happen (inability to resolve addresses that were "negatively" cached):
		// - before starting kwt, resolve 'foo.test'
		//   - mDNSResponder will cache negative result with a very high TTL
		//     because foo.test isnt typically resolvable
		// - start kwt net start --dns-map test=127.0.0.1
		// - resolve 'foo.test' again, expecting 127.0.0.1
		//   - via dig it works because it bypasses OS X resolution
		//   - via curl it does not work since negative result is still cached by OS X
		// See mDNSResponder's internal cache via:
		// $ log stream --predicate 'process == "mDNSResponder"' --info
		// $ sudo killall -INFO mDNSResponder
		c.flushOSX()

	default:
		c.logger.Debug(c.logTag, "Skipping clearing of OS DNS cache")
	}
}

func (c DNSOSCache) flushOSX() {
	out, err := exec.Command("killall", "-HUP", "mDNSResponder").CombinedOutput()
	if err != nil {
		c.logger.Debug(c.logTag, "Failed clearing mDNSResponder cache: %s (output: %s)", err, out)
	} else {
		c.logger.Debug(c.logTag, "Successfully cleared via mDNSResponder")
		return
	}

	// Try flushing Directory Service cache which may on some versions of OS X do the trick
	out, err = exec.Command("discoveryutil", "udnsflushcaches").CombinedOutput()
	if err != nil {
		c.logger.Debug(c.logTag, "Failed clearing via discoveryutil: %s (output: %s)", err, out)
	} else {
		c.logger.Debug(c.logTag, "Successfully cleared via discoveryutil")
	}
}
