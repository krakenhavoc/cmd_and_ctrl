package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kotis, Sibsig Champion — Legendary Creature — Zombie Warrior {B}{G}{U},
// 3/3 (Tarkir: Dragonstorm Commander):
//
//	"Once during each of your turns, you may cast a creature spell from your
//	 graveyard by exiling three other cards from your graveyard in addition
//	 to paying its other costs.
//	 Whenever one or more creatures you control enter, if one or more of them
//	 entered from a graveyard or was cast from a graveyard, put two +1/+1
//	 counters on Kotis."
//
// The first ability is Gisa and Geralf's standing graveyard permission (one
// creature cast per turn per Kotis object, your turns only, the spell's costs
// paid as usual) with an extra price: three OTHER cards of the graveyard
// exiled. That is AltCostKeyGraveyardExile, escape's exile cost under a key of
// its own, so the cast is neither an escape nor a replacement of the mana
// cost. The second is Celes, Rune Knight's batch trigger (entered from a
// graveyard, or Provenance.FromZone is the graveyard), with Kotis itself
// counted among the creatures and the condition judged for the whole batch.
//
// Seam: AltCostKeyGraveyardExile.
func init() {
	Register(Spec{
		OracleID:     "0177b410-b559-491f-b393-ac3ed774653c",
		Name:         "Kotis, Sibsig Champion",
		Completeness: CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:                    game.ZoneGraveyard,
			CastOnly:                true,
			Filter:                  game.PermissionFilter{CreatureOnly: true},
			PerType:                 []string{game.PermissionTypeCreature},
			Timing:                  game.TimingYourTurnOnly,
			AltCostKey:              game.AltCostKeyGraveyardExile,
			ExileOtherFromGraveyard: 3,
			Label:                   "Once during each of your turns, cast a creature spell from your graveyard by exiling three other cards from your graveyard (Kotis, Sibsig Champion)",
		}},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventETB, cmdrB13CreatureFromGraveyard,
				"Kotis, Sibsig Champion — put two +1/+1 counters on Kotis",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 2}))),
		},
	})
}
