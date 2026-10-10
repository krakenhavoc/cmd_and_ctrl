package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urza, Lord Protector — Legendary Creature — Human Artificer {1}{W}{U},
// 2/4:
//
//	"Artifact, instant, and sorcery spells you cast cost {1} less to cast.
//	 {7}: If you both own and control Urza, Lord Protector and an artifact
//	 named The Mightstone and Weakstone, exile them, then meld them into
//	 Urza, Planeswalker. Activate only as a sorcery."
//
// The first line is an ordinary CR 601.2f reduction on the controller's
// own artifact, instant and sorcery spells.
//
// The second is the meld ability (CR 701.42a, 712.5e; ADR 0145, #2699).
// Its condition is checked as it resolves (MeldWith): Urza still on the
// battlefield and an artifact named The Mightstone and Weakstone beside
// him, both owned and controlled by the activator. Then both are exiled
// together and come back as one permanent, Urza, Planeswalker, which has
// only the combined back face's characteristics (CR 712.8g) and whose
// own abilities are in urza_planeswalker.go. If Urza is the commander,
// so is the melded permanent (CR 903.3b). A Clone named Urza, or a copy
// of the Mightstone, is exiled but can't be melded and stays in exile
// (CR 701.42c).
func init() {
	Register(Spec{
		OracleID:     "df2af646-3e5b-43a3-8f3e-50565889f456",
		Name:         "Urza, Lord Protector",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Artifact, instant, and sorcery spells you cast cost {1} less to cast.",
				YourSpell(), ArtifactInstantOrSorcerySpell()),
		},
		Activated: []ActivatedAbility{{
			Label:        "{7}: If you both own and control Urza, Lord Protector and an artifact named The Mightstone and Weakstone, exile them, then meld them into Urza, Planeswalker. Activate only as a sorcery.",
			Cost:         ManaCost("{7}"),
			SorcerySpeed: true,
			Effect:       Do(MeldWith{Partner: "The Mightstone and Weakstone", PartnerType: "artifact"}),
		}},
	})
}
