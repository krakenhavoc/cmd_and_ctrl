package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gunner Conscript — Creature — Human Mercenary {1}{G}, 2/2 (Fallout,
// #5948):
//
//	"Trample
//	 This creature gets +1/+1 for each Aura and Equipment attached to it.
//	 When this creature dies, if it was enchanted, create a Junk token.
//	 When this creature dies, if it was equipped, create a Junk token."
//
// The pump is a layer 7c self-modification counted on every recompute over
// the Auras and Equipment attached to it. The two printed dies triggers are
// two rows (CR 603.2c: an enchanted, equipped Conscript makes two Junk).
// Each is an intervening "if" (CR 603.4) read against the departure record
// of what was attached as it died (gapB62AttachedTo): an Aura put into the
// graveyard and an Equipment that has become unattached by the time the
// trigger is looked at still count.
//
// No simplification.
func init() {
	junk := Do(CreateToken{Template: JunkToken(), N: 1})
	Register(Spec{
		OracleID:        "f3d3ed94-f32c-4a67-8e2a-765e454a80df",
		Name:            "Gunner Conscript",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label:     "Gets +1/+1 for each Aura and Equipment attached to it.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfCreature,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, a := range g.BattlefieldCardsForEffect() {
					if a.IsAttachedTo(source.InstanceID) && (a.IsAura() || a.HasSubtype("Equipment")) {
						n++
					}
				}
				c.Power += n
				c.Toughness += n
			},
		}},
		Triggered: []game.TriggeredAbility{
			gapB62DiesAttached("Gunner Conscript — it was enchanted: create a Junk token", true, junk),
			gapB62DiesAttached("Gunner Conscript — it was equipped: create a Junk token", false, junk),
		},
	})
}
