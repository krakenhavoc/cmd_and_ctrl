package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doom Weaver — {4}{B}{B} Creature — Spider Horror 1/8
// (Innistrad: Crimson Vow Commander, #5675):
//
//	"Reach
//	 Soulbond (You may pair this creature with another unpaired creature
//	 when either enters. They remain paired for as long as you control
//	 both of them.)
//	 As long as Doom Weaver is paired with another creature, each of
//	 those creatures has "When this creature dies, draw cards equal to
//	 its power.""
//
// Soulbond is the keyword's two triggers (CR 702.95a, soulbond.go). The
// last line is a layer-6 grant to both creatures of the pair. A dies
// trigger looks back at the creature as it last existed (CR 603.10a):
// the two were still paired then, so whichever dies has the ability,
// and if both die together both do. "Its power" is that creature's
// power as it last existed on the battlefield (CR 608.2h), and the
// player who controlled it draws.
//
// No simplification.
const doomWeaverGrant = "doom-weaver/draw"

func init() {
	Register(Spec{
		OracleID:        "172b94a0-2801-4623-aed8-902591c2d591",
		Name:            "Doom Weaver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", game.KeywordSoulbond},
		Triggered:       Soulbond("Doom Weaver"),
		Grants: []AbilityGrant{{
			Key: doomWeaverGrant,
			Triggered: []game.TriggeredAbility{
				WhenThisDies("Doom Weaver — draw cards equal to this creature's power", doomWeaverDraw),
			},
			Text: "When this creature dies, draw cards equal to its power.",
		}},
		Static: []game.StaticAbility{
			PairedCreaturesHave("As long as Doom Weaver is paired with another creature, each of those creatures has \"When this creature dies, draw cards equal to its power.\"",
				doomWeaverGrant),
		},
	})
}

// doomWeaverDraw draws cards equal to the dead creature's last-known
// power. A power of zero or less draws nothing.
func doomWeaverDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.TriggeringPermanent()
	if !ok || info.Power <= 0 {
		return nil
	}
	return DrawCards{Player: item.Controller, N: info.Power}.Apply(ctx)
}
