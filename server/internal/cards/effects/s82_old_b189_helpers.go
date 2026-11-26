package effects

// s82_old_b189_helpers.go — shared vocabulary for the S82 older-sets batch 189
// cards (#5525).

// ExileNonlandFromHandOrGraveyard is "target opponent reveals their hand. You
// choose a nonland card from that player's graveyard or hand and exile it"
// (Memory Leak, Never Happened): the hand is revealed and the card is chosen
// from the revealed hand and the graveyard together; a land card is never a
// choice from either. It is exiled, not discarded.
func ExileNonlandFromHandOrGraveyard(ctx *Context) error {
	return ChooseFromRevealedHand{
		Player:          TargetedPlayer(ctx),
		Filter:          Nonland(),
		Label:           "nonland card",
		Exile:           true,
		FromGraveyard:   true,
		GraveyardFilter: Nonland(),
	}.Apply(ctx)
}
