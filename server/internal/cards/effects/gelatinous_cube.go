package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gelatinous Cube — Creature — Ooze {2}{B}{B}, 4/3 (Adventures in the Forgotten
// Realms):
//
//	"Engulf — When this creature enters, exile target non-Ooze creature an
//	 opponent controls until this creature leaves the battlefield.
//	 Dissolve — {X}{B}: Put target creature card with mana value X exiled
//	 with this creature into its owner's graveyard."
//
// Both halves are Bronzebeak Foragers' shape: the exile is ExileUntil with
// ThisLeaves and the CR 607.2a link stamped (CR 610.3), and the activated
// ability targets a card exiled with this very object whose mana value is
// the X paid, re-checked as it resolves (CR 608.2b). A Cube gone before the
// trigger resolves exiles nothing (CR 610.3b). "Engulf" and "Dissolve" are
// ability words with no rules meaning. No simplification.
func init() {
	enters := Targeting(
		WhenThisEnters("Gelatinous Cube — exile target non-Ooze creature an opponent controls until this creature leaves the battlefield",
			gelatinousCubeExile),
		TargetCreature("target non-Ooze creature an opponent controls", OpponentControls(), Not(Subtype("Ooze"))))
	Register(Spec{
		OracleID:     "3c2e2e69-cd44-458f-a5b0-8591baa24e65",
		Name:         "Gelatinous Cube",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
		Activated: []ActivatedAbility{{
			Label:   "{X}{B}: Put target creature card with mana value X exiled with this creature into its owner's graveyard.",
			Cost:    ManaCost("{X}{B}"),
			Targets: ExiledWithThis(TargetCardInExile("target creature card with mana value X exiled with this creature", Creature()).WithManaValueEqualsX()),
			Effect:  gelatinousCubeDissolve,
		}},
	})
}

func gelatinousCubeExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (ExileUntil{
			Target:     t.ID,
			ThisLeaves: true,
			Linked:     true,
			Label:      "Gelatinous Cube — returns when Gelatinous Cube leaves the battlefield",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

func gelatinousCubeDissolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		_, err := cmdr2S32PutExiledCardIntoGraveyard(g, t.ID)
		return err
	}
	return nil
}
