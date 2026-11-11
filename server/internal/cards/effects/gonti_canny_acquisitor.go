package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gonti, Canny Acquisitor — Legendary Creature — Aetherborn Rogue
// {2}{B}{G}{U}, 5/5 (Outlaws of Thunder Junction Commander):
//
//	"Spells you cast but don't own cost {1} less to cast.
//	 Whenever one or more creatures you control deal combat damage to a
//	 player, look at the top card of that player's library, then exile it
//	 face down. You may play that card for as long as it remains exiled, and
//	 mana of any type can be spent to cast that spell."
//
// The cost reduction is a CostModifier over the cast spell's owner. The
// trigger is one per player the controller's creatures connect with in a
// combat damage step (CR 603.2c), and the exile is Gonti, Night Minister's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7923ef4a-b1be-4f6a-b17b-4eab101390de",
		Name:         "Gonti, Canny Acquisitor",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Spells you cast but don't own cost {1} less to cast.",
				YourSpell(), cmdr2S45SpellYouDontOwn()),
		},
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Gonti, Canny Acquisitor — look at the top card of that player's library, then exile it face down; you may play it",
				cmdr2S45ExileTopFaceDownToPlay),
		},
	})
}
