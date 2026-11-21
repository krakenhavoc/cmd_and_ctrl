package effectstest

// counter_unless_paid_exile_test.go — #5236: "counter target spell
// unless its controller pays {N}. If that spell is countered this way,
// exile it instead of putting it into its owner's graveyard" (CR
// 118.12a, CR 701.6a, CR 614.1a), with Syncopate (#4462) and No More
// Lies (#4469) on top.

const (
	SyncopateOracle  = "318f7f70-e374-40ef-8afb-3389c10461d8"
	NoMoreLiesOracle = "ead1ee6a-e0da-46d5-89e8-31e8c0270bc2"

	// counterExileKey is the engine's keyed decline for a counter-unless
	// that exiles: the destination rides the prompt as this key.
	CounterExileKey = "pay-unless/counter-guarded-exile"
)
