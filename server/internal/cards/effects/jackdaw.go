package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jackdaw — Legendary Artifact — Vehicle, {1}{U}{R}:
//
//	"Whenever Jackdaw deals combat damage to a player, you may discard
//	 your hand. If you do, draw a card for each artifact you control.
//	 Crew 3"
//
// The "if you do" half is a CR 603.12 REFLEXIVE trigger, exactly the
// Undead Butler / Generous Plunderer shape: the "you may" belongs to
// the parent (Optional wraps the whole combat-damage trigger), and
// once you've said yes the discard always happens — the reflexive
// half that draws is then unconditional, because "if you do" is
// satisfied by having done it, empty hand or not.
//
// The count is artifacts controlled AFTER the discard resolves, as
// the brief specifies — jackdawDrawForArtifacts reads the board at
// the moment the reflexive trigger itself resolves, which is always
// after the discard that created it. Jackdaw counts itself: nothing
// on the card excludes it, and it is still on the battlefield (the
// ability doesn't sacrifice it).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac161100-46b7-4f0f-a72b-84aaa45980ad",
		Name:         "Jackdaw",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(
				WheneverThisDealsCombatDamageToAPlayer("Jackdaw — discard your hand", jackdawDiscardHand),
				"Jackdaw — discard your hand? (Draw a card for each artifact you control.)"),
		},
		Activated: []ActivatedAbility{{
			Label:  "Crew 3",
			Cost:   CrewCost(3),
			Effect: CrewEffect("Jackdaw"),
		}},
	})
}

// jackdawDiscardHand is the parent trigger's body: discard the whole
// hand (no choice of what to pitch — it's all of it), then create the
// CR 603.12 reflexive trigger that draws. The discard happened the
// instant this Effect ran (Optional's "you may" already answered
// yes), so the reflexive half is unconditional.
func jackdawDiscardHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if _, err := discardWholeHand(g, item.Controller); err != nil {
		return err
	}
	return ReflexiveTrigger{
		Label:  "Jackdaw — draw a card for each artifact you control",
		Effect: jackdawDrawForArtifacts,
	}.Apply(ctx)
}

// jackdawDrawForArtifacts is the reflexive half: draw a card for each
// artifact the controller controls, counted when THIS trigger
// resolves (after the discard, per the brief).
func jackdawDrawForArtifacts(g *game.Game, item *game.StackItem) error {
	n := b03ArtifactsControlled(g, item.Controller)
	if n == 0 {
		return nil
	}
	return g.DrawNForEffect(item.Controller, n)
}
