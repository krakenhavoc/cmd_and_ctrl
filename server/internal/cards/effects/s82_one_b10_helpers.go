package effects

// s82OneB10OilCounters reads the oil counters on the permanent an ability
// came from as it resolves: live while it is still the same object on the
// battlefield, and as it last existed there once it has gone (CR 608.2h).
func S82OneB10OilCounters(ctx *Context) int {
	info, ok := ctx.SourcePermanent()
	if !ok {
		return 0
	}
	return info.Counters[OilCounter]
}

// oilCounter is the oil counter kind (CR 122.1), the on-disk name Vat of
// Rebirth and Mindsplice Apparatus use.
const OilCounter = "oil"
