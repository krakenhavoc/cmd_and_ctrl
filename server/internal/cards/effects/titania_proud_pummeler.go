package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Titania, Proud Pummeler — Legendary Creature — Human Villain {3}{R},
// 3/3 (Marvel Super Heroes Commander, #4455, seam #4722):
//
//	"First strike
//	 Melee (Whenever this creature attacks, it gets +1/+1 until end of
//	 turn for each opponent you attacked this combat.)
//	 Other creatures you control have melee."
//
// Melee is the engine's derived keyword trigger (CR 702.121, game/melee.go):
// one trigger per instance on the effective ability list, the opponent count
// read as each resolves. The grant is an ordinary layer-6 keyword grant over
// the other creatures you control (Tyranid Prime's shape), so a creature
// that prints melee has two instances under Titania and each triggers on its
// own (CR 702.121b). Titania carries her own printed melee and not the
// granted one: the grant says "other".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6d51f00e-46a5-4064-a0c8-e840c66dbf2b",
		Name:            "Titania, Proud Pummeler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", game.KeywordMelee},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Others: true, YoursOnly: true}, game.KeywordMelee),
		},
	})
}
