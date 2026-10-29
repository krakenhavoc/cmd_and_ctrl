package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jennifer Walters // The Sensational She-Hulk — a transforming
// double-faced legendary creature (Marvel Super Heroes, #4455):
//
//	Jennifer Walters — Legendary Creature — Human Advisor Hero
//	  "Your opponents can't cast spells during your turn.
//	   {3}{G}{W}{W}: Transform Jennifer Walters. Activate only as a
//	   sorcery."
//	The Sensational She-Hulk — Legendary Creature — Gamma Hero
//	  "Reach, trample
//	   Your opponents can't cast spells during your turn.
//	   Whenever a creature you control is dealt damage, you may have The
//	   Sensational She-Hulk deal that much damage to any target. Do this
//	   only once each turn."
//
// Voice of Victory's cast restriction on both faces; the transform is Aang's
// (TransformThis, CR 712) at sorcery speed. The reflection is Screaming
// Nemesis's shape: the trigger waits for the damage to settle (ThatMuchDamage,
// so simultaneous damage from several sources is one total, CR 120.3), with the target chosen as the trigger goes on the stack and the
// "you may" asked as it resolves. "Only once each turn" limits the DOING
// (Baron Strucker's b28DoneThisTurn), so declining does not use the turn up,
// and a second trigger from the same combat is dropped once the first was
// carried out. She-Hulk deals the damage herself; if she has left, the
// damage is still dealt as she last existed (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     jenniferWaltersOracle,
		Name:         "Jennifer Walters",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			OpponentsCantCastDuringYourTurn("Jennifer Walters — your opponents can't cast spells during your turn.", nil),
		},
		Activated: []ActivatedAbility{{
			Label:        "{3}{G}{W}{W}: Transform Jennifer Walters. Activate only as a sorcery.",
			Cost:         ManaCost("{3}{G}{W}{W}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:        jenniferWaltersOracle + "#1",
		Name:            "The Sensational She-Hulk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		CastRestrictions: []game.CastRestriction{
			OpponentsCantCastDuringYourTurn("The Sensational She-Hulk — your opponents can't cast spells during your turn.", nil),
		},
		Triggered: []game.TriggeredAbility{
			Optional(withAtBatchEnd(Targeting(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := b28CreatureWasDealtDamage(ev, g)
				return ok && c.Controller == source.Controller && !b28DoneThisTurn(g, source.InstanceID, sheHulkDone)
			}, sheHulkLabel, sheHulkReflect), TargetAny()), ThatMuchDamage()),
				"The Sensational She-Hulk — deal that much damage to any target?"),
		},
	})
}

const (
	jenniferWaltersOracle = "c867a778-97ef-4872-b890-b088b96a1281"
	sheHulkLabel          = "The Sensational She-Hulk — deal that much damage to any target"
	sheHulkDone           = "The Sensational She-Hulk — reflected damage this turn"
)

// sheHulkReflect deals the damage the creature was dealt to the chosen
// target, once each turn.
func sheHulkReflect(g *game.Game, item *game.StackItem) error {
	if b28DoneThisTurn(g, item.SourceCardID, sheHulkDone) {
		return nil
	}
	ctx := NewContext(g, item)
	ts := ctx.LegalTargets()
	if len(ts) == 0 {
		return nil
	}
	b28NoteDoneThisTurn(g, item.SourceCardID, sheHulkDone)
	return DealDamage{Source: item.SourceCardID, Target: ts[0].ID, Amount: settledEventAmount(item)}.Apply(ctx)
}
