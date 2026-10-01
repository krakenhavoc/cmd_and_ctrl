package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Renegade Krasis — Creature — Beast Mutant {1}{G}{G}, 3/2:
//
//	"Evolve (Whenever a creature you control enters, if that creature
//	 has greater power or toughness than this creature, put a +1/+1
//	 counter on this creature.)
//	 Whenever this creature evolves, put a +1/+1 counter on each other
//	 creature you control with a +1/+1 counter on it."
//
// #1805's "evolves" proof card (ADR 0106 §3, owner decision 5). The
// creature evolves when its evolve ability puts one or more +1/+1
// counters on it (CR 702.100b), which the engine reports as
// EventEvolved (game/evolve.go). A counter from anything else — a
// proliferate, a Hardened Scales on some other placement — is not an
// evolution and does not trigger this.
//
// "Each other creature you control with a +1/+1 counter on it" is read
// as the trigger resolves, so a creature that got its first counter in
// response counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7b7e622c-815a-44a0-88ff-62bbe8dd5582",
		Name:            "Renegade Krasis",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordEvolve},
		Triggered: []game.TriggeredAbility{
			WhenThisEvolves("Renegade Krasis — put a +1/+1 counter on each other creature you control with one", renegadeKrasisSpread),
		},
	})
}

// renegadeKrasisSpread puts a +1/+1 counter on each other creature the
// controller controls that has one.
func renegadeKrasisSpread(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item).asGroupMember()
	// "Other" is this object. A Krasis that left and came back is a new
	// object (CR 400.7) and is one of the others.
	selfStillHere := !sourceIsNewObject(g, item)
	for _, c := range g.BattlefieldCardsForEffect() {
		if (selfStillHere && c.InstanceID == item.SourceCardID) || c.Controller != item.Controller || !c.IsCreature() || c.Counters[game.CounterPlusOne] <= 0 {
			continue
		}
		if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
