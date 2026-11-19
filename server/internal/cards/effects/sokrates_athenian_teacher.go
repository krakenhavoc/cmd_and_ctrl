package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sokrates, Athenian Teacher — Legendary Creature — Human Advisor {1}{W}{U},
// 0/4 (Assassin's Creed, #5948):
//
//	"Defender
//	 Sokrates has hexproof as long as it's untapped.
//	 Sokratic Dialogue — {T}: Until end of turn, target creature gains "If
//	 this creature would deal combat damage to a player, prevent that damage.
//	 This creature's controller and that player each draw half that many cards,
//	 rounded down.""
//
// The hexproof is Karakyk Guardian's shape on a different gate: a layer 6
// grant to Sokrates alone, on while he is untapped, read on every recompute.
//
// The granted ability is a prevention shield on the target creature as the
// source, until end of turn (CR 615): "all combat damage it would deal to a
// player" (Maze of Ith's source pin with the players recipient set, read as the
// damage would be dealt), with the CR 615.5 follow-up Vengeful Archon's shield
// carries. The follow-up is owed once per damage instance with the amount
// actually prevented, and draws half of it rounded down for the creature's
// controller as it was when the damage would have been dealt (last-known
// information) and for the damaged player. Damage to a creature or planeswalker,
// or damage that can't be prevented, is not changed (nothing is prevented, so
// nothing is drawn).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f4a9d567-513b-487e-afed-c6e6513a7862",
		Name:            "Sokrates, Athenian Teacher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{{
			Label: "Has hexproof as long as it's untapped.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && !source.Tapped
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				appendKeywordsTo(c, []string{"hexproof"})
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "Sokratic Dialogue — {T}: Until end of turn, target creature gains \"If this creature would deal combat damage to a player, prevent that damage. This creature's controller and that player each draw half that many cards, rounded down.\"",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect:  sokratesDialogue,
		}},
	})
}

func sokratesDialogue(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok || !onBattlefield(g, id) {
		return nil
	}
	ref, zone, ok := g.DamageSourceRefLocked(id)
	if !ok {
		return nil
	}
	g.PreventDamageFromSourceThisTurnForEffect(game.DamageShield{
		EffectSource: item.SourceCardID,
		Controller:   item.Controller,
		Source:       ref,
		SourceZone:   zone,
		Recipients:   game.DamageRecipientFilter{Players: true},
		CombatOnly:   true,
		Then:         sokratesDrawHalfBody,
		Label:        "Sokrates, Athenian Teacher — prevent its combat damage to a player; each draws half that many cards",
	})
	return nil
}

// sokratesDrawHalfBody is the shield's CR 615.5 follow-up: the damaged player
// and the damaging creature's controller each draw half the damage prevented,
// rounded down. An on-disk key.
var sokratesDrawHalfBody = game.DelayedBody("sokrates/draw-half-prevented", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	n := p.Amount / 2
	if n <= 0 || item.Trigger == nil {
		return nil
	}
	if err := g.DrawNForEffect(p.Player, n); err != nil {
		return err
	}
	return g.DrawNForEffect(item.Trigger.Event.Target, n)
})
