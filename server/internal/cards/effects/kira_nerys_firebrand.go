package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kira Nerys, Firebrand — {1}{R}{W} Legendary Creature — Bajoran Soldier
// 4/3 (Star Trek, #4973):
//
//	"Trample
//	 If a creature you control would enter while you control an attacking
//	 creature, instead it enters tapped and attacking.
//	 Whenever another creature you control enters, it gets +1/+1 until
//	 end of turn."
//
// The middle line is a CR 614.1d replacement on every creature entering
// under Kira's controller, checked against the creature as it would exist
// on the battlefield (CR 614.12). Its CR 508.4 question — which player,
// planeswalker or battle the creature attacks — is the engine's
// entry_attack_target prompt, asked before the creature enters
// (CR 614.12a; game/entry_attack.go). The creature was put onto the
// battlefield attacking, never declared, so it triggers no "whenever ~
// attacks" ability (CR 508.4). The condition is "while you control an
// attacking creature", Kira herself included, read as the creature would
// enter.
//
// The last line is an ordinary enters trigger: "it" is the creature that
// entered, named by the trigger's event.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ff9f5bbb-43c4-481e-9bd7-db9f246355ae",
		Name:            "Kira Nerys, Firebrand",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			YourCreaturesEnterTappedAndAttackingWhileYouAttack("Kira Nerys, Firebrand"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AnotherCreatureEnteredUnderYourControl,
				"Kira Nerys, Firebrand — it gets +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return BoostUntilEOT{
						Target: item.Trigger.Event.CardID,
						Power:  1, Toughness: 1,
						Label: "Kira Nerys, Firebrand — +1/+1 until end of turn",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// YourCreaturesEnterTappedAndAttackingWhileYouAttack is "If a creature you
// control would enter while you control an attacking creature, instead it
// enters tapped and attacking" (CR 614.1d), with CR 508.4's choice of what
// it attacks asked of its controller as it enters (game.EntryAttackChoice).
// `name` heads the prompt and the CR 616 label.
func YourCreaturesEnterTappedAndAttackingWhileYouAttack(name string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield ||
				ev.OldZone == game.ZoneBattlefield {
				return false
			}
			if g.EntryWouldBeControllerForEffect(ev) != src.Controller || !controlsAnAttackingCreature(g, src.Controller) {
				return false
			}
			return g.EntryWouldBeCreatureForEffect(ev)
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		EntryAttack: &game.EntryAttackChoice{Question: name + " — what does it attack?"},
		Label:       name + ": enters tapped and attacking",
	}
}

// controlsAnAttackingCreature reports whether `player` controls a creature
// that is attacking.
func controlsAnAttackingCreature(g *game.Game, player uuid.UUID) bool {
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == player && c.AttackingTarget != uuid.Nil && c.IsCreature() {
			return true
		}
	}
	return false
}
