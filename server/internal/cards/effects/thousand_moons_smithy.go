package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const thousandMoonsSmithyOracleID = "32af5f7b-a970-484a-9aff-226749551d32"

// Thousand Moons Smithy — Legendary Artifact {2}{W}{W}, the front of a
// transforming card whose back is Barracks of the Thousand (The Lost
// Caverns of Ixalan, #4470):
//
//	"When Thousand Moons Smithy enters, create a white Gnome Soldier
//	 artifact creature token with 'This token's power and toughness are
//	 each equal to the number of artifacts and/or creatures you control.'
//	 At the beginning of your first main phase, you may tap five untapped
//	 artifacts and/or creatures you control. If you do, transform
//	 Thousand Moons Smithy."
//
// The tap is Caparocti Sunborn's: a card-set pick of exactly five (or none,
// the "you may" declined) made as the trigger resolves, each re-read before
// it is tapped, and the transform happens only if all five were tapped
// ("if you do"). The Smithy itself may be one of the five.
//
// Barracks of the Thousand:
//
//	"{T}: Add {W}.
//	 Whenever you cast an artifact or creature spell using mana produced by
//	 Barracks of the Thousand, create a white Gnome Soldier …"
//
// The trigger is a mana spend rider on the {T} ability (as Tecutlan's):
// it goes on the stack above the spell that caused it, so it resolves
// first (the ruling), and only an artifact or creature spell paid with
// that mana counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     thousandMoonsSmithyOracleID,
		Name:         "Thousand Moons Smithy",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Thousand Moons Smithy — create a Gnome Soldier artifact creature token",
				Do(CreateToken{Template: GnomeSoldierToken(), N: 1})),
			AtYourStep(game.StepPrecombatMain, "Thousand Moons Smithy — you may tap five untapped artifacts and/or creatures you control; if you do, transform it",
				thousandMoonsSmithyTapFive),
		},
	})
	Register(Spec{
		OracleID:       thousandMoonsSmithyOracleID + "#1",
		Name:           "Barracks of the Thousand",
		Completeness:   CompletenessFull,
		Purpose:        game.Purpose{Tokens: 1, CreatureTokens: 1},
		ReadsManaSpent: game.ReadsPoolManaOnly,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "{T}: Add {W}",
			SpendRiders: []game.ManaSpendRider{WhenManaSpent("Barracks of the Thousand", game.ManaSpendTrigger{
				Label:  "Barracks of the Thousand — create a Gnome Soldier artifact creature token",
				Effect: Do(CreateToken{Template: GnomeSoldierToken(), N: 1}),
			}, ManaRestrictCast, ManaRestrictAnyType("Artifact", "Creature"))},
		}},
	})
}

func thousandMoonsSmithyTapFive(g *game.Game, item *game.StackItem) error {
	candidates := caparoctiUntappedArtifactsOrCreatures(g, item.Controller)
	if len(candidates) < 5 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Thousand Moons Smithy — tap five untapped artifacts and/or creatures you control to transform it?",
		Cards:    candidates,
		Min:      0,
		Max:      5,
		ThenKey:  thousandMoonsSmithyTap,
		LiveZone: game.ZoneBattlefield,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// thousandMoonsSmithyTap taps the five, then transforms. "Exactly five" is
// a registered set rule, not a closure.
var thousandMoonsSmithyTap = askedPickRule("choose-cards/thousand-moons-smithy-tap-five",
	func(picked []game.Card) bool { return len(picked) == 5 },
	func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) != 5 {
			return nil
		}
		still := caparoctiUntappedArtifactsOrCreatures(ctx.Game, ctx.Controller())
		tapped := 0
		for _, id := range r.Picked {
			if !slices.Contains(still, id) {
				continue
			}
			if err := ctx.Game.TapTargetForEffect(id); err != nil {
				return err
			}
			tapped++
		}
		if tapped != 5 {
			return nil
		}
		return TransformThis{}.Apply(ctx)
	})
