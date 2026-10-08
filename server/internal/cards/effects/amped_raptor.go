package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Amped Raptor — Creature — Dinosaur {1}{R}, 2/1:
//
//	"First strike
//	 When this creature enters, you get {E}{E} (two energy counters).
//	 Then if you cast it from your hand, exile cards from the top of
//	 your library until you exile a nonland card. You may cast that card
//	 by paying an amount of {E} equal to its mana value rather than
//	 paying its mana cost."
//
// The energy comes first, so the two it gives can pay for the hit. "Then
// if you cast it from your hand" is checked as the trigger resolves (it
// is not an intervening if), off the cast event (b30CastFromHand), so a
// reanimated or blinked Raptor only gets the energy.
//
// The cast is a per-card CastPermission (ADR 0066) priced by
// EnergyEqualToManaValue (ADR 0129 §5): Bolas's Citadel's "rather than
// pay its mana cost" with energy for life, so it is the only way to cast
// the card this way, a caster short of the energy cannot claim it (CR
// 118.3), and an {X} in the card's cost is 0 (CR 107.3b). The rulings
// put the cast inside the resolution with its type's timing ignored, so
// the grant is Fevered Suspicion's: flash timing (CR 608.2g), and a
// window that closes when you pass priority, leaving an uncast card in
// exile.
//
// SANDBOX SIMPLIFICATION, cascade's (cascade.go): the cast is made right
// after the trigger finishes resolving rather than during it. The
// pass-closed window keeps it no stronger than printed.
func init() {
	Register(Spec{
		OracleID:        "3ed14ea7-ee62-4604-af3b-deafb00108c3",
		Name:            "Amped Raptor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Purpose:         game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Amped Raptor — you get {E}{E}, then exile until a nonland card", ampedRaptorEnters),
		},
	})
}

// ampedRaptorAltCostKey is the claim the permission synthesises. An
// on-disk identity (it lands on StackItem.AltCost): never renamed.
const ampedRaptorAltCostKey = "amped_raptor"

func ampedRaptorEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (GetEnergy{N: 2}).Apply(ctx); err != nil {
		return err
	}
	if !b30CastFromHand(g, item.SourceCardID) {
		return nil
	}
	return exileUntilNonlandGrantingCast(ctx, game.CastPermission{
		AltCostKey:             ampedRaptorAltCostKey,
		EnergyEqualToManaValue: true,
		Timing:                 game.TimingFlash,
		CastOnly:               true,
		Duration:               ctx.Game.UntilEndOfTurnDuration(),
		LapseOnPass:            game.LapseStaysInExile,
		Source:                 ctx.Source(),
		SourceName:             "Amped Raptor",
		Label:                  "Pay {E} equal to its mana value (Amped Raptor)",
	})
}
