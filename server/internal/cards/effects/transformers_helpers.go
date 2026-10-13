package effects

// transformers_helpers.go — what the More Than Meets the Eye cards (ADR
// 0152, #3195) share beyond convert.go: the "convert it. If you do, …"
// reading.

// convertThisReporting is "convert <this>" for a clause that goes on
// "If you do, …": it runs ConvertThis and reports whether the source
// actually turned over. A convert that does nothing — the permanent
// already converted or transformed since the ability went on the stack
// (CR 701.28e), or it cannot convert (CR 701.28c, 701.28d, 701.28f) —
// is not "doing it", so the sentence after it must not happen.
func convertThisReporting(ctx *Context) (bool, error) {
	before, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	if !ok {
		return false, nil
	}
	if err := (ConvertThis{}).Apply(ctx); err != nil {
		return false, err
	}
	after, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	return ok && after.TurnOvers > before.TurnOvers, nil
}
