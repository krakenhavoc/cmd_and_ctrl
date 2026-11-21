package effectstest

// draw_payoffs_test.go — the cards that turn "I drew a card" into
// something else. The shared assertion across all of them is that the
// trigger fires PER CARD, because EventDrawCard does: a spell that
// draws three has to drain for three, make three Insects, or deal
// three damage. A payoff that fires once per spell is a different and
// much worse card, and it is the easy bug to write.

const (
	PsychosisCrawlerOracle = "2876e74f-a242-4995-9702-0b737a1ab67a"
	LocustGodOracle        = "e025a714-02da-4b0c-8021-cf3e8dc9b19e"
	HowlingMineOracle      = "d26b27db-a567-4631-b4b6-7294222fbdd1"
	HedronCrabOracle       = "7216f974-3c84-40ef-b904-82019900c204"
)
