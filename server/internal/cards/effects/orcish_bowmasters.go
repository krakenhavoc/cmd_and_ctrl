package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Orcish Bowmasters — Creature — Orc Archer {1}{B}, 1/1 (EDHREC rank
// 257):
//
//	"Flash
//	 When this creature enters and whenever an opponent draws a card
//	 except the first one they draw in each of their draw steps, this
//	 creature deals 1 damage to any target. Then amass Orcs 1."
//
// The card #1236 was filed for, and the reason it stayed unregistered:
// dropping "amass Orcs 1" would have lost the whole army-building half
// of a two-mana creature, which is not a caveat, it is a different
// card. Amass is a primitive now (amass.go, ADR 0087) and the Orc Army
// is the token the keyword mints.
//
// ONE printed ability with TWO trigger conditions, not two abilities:
// `OnAny` over EventETB and EventDrawCard with a predicate that knows
// which is which. Two declarations would queue a CR 603.3b ordering
// prompt whenever both could fire and would ask for one target each.
//
// Flash rides PrintedKeywords, which is what makes the Bowmasters a
// wheel's instant-speed punish: it enters in response to the draw
// spell and its ETB half fires before the wheel resolves.
//
// Sandbox simplification, declared — it is Notion Thief's, word for
// word. "Except the first one they draw in each of their draw steps"
// is read as "except any draw during their own draw step", because the
// engine keeps no per-draw-step tally (`drawnInOwnDrawStep`,
// draw_replacements.go). A second draw in an opponent's draw step —
// an instant they cast there, a draw-step trigger — does not fire the
// Bowmasters. Weaker than printed for the Bowmasters' controller,
// never stronger, and closing it is the per-draw-step tally on the
// "Draw-replacement count" seam row.
func init() {
	Register(Spec{
		OracleID:        "ea5103f5-27e0-4eb1-902c-7f34652d6bf3",
		Name:            "Orcish Bowmasters",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"Every card an opponent draws during their own draw step is left alone, not just the first one."},
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				OnAny([]game.EventKind{game.EventETB, game.EventDrawCard},
					orcishBowmastersFires,
					"Orcish Bowmasters — 1 damage to any target, then amass Orcs 1",
					orcishBowmastersShoot),
				TargetAny()),
		},
	})
}

// orcishBowmastersFires is the printed condition, both halves: the
// Bowmasters entering, or an opponent drawing outside their own draw
// step.
func orcishBowmastersFires(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
	switch ev.Kind {
	case game.EventETB:
		return Self(ev, source, ch, g)
	case game.EventDrawCard:
		return ByAnOpponent(ev, source, ch, g) && !drawnInOwnDrawStep(g, ev.Actor)
	}
	return false
}

// orcishBowmastersShoot is "this creature deals 1 damage to any
// target. Then amass Orcs 1."
//
// The amass is unconditional on the damage landing — "then", not "if
// you do" — but a trigger whose only target became illegal never
// resolves at all (CR 608.2b), so the Army is not amassed either, and
// that is the printed behaviour rather than an omission.
func orcishBowmastersShoot(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if targets := ctx.LegalTargets(); len(targets) > 0 {
		if err := (DealDamage{Source: item.SourceCardID, Target: targets[0].ID, Amount: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return Amass{Subtype: "Orc", N: 1}.Apply(ctx)
}
