package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Insidious Roots — Enchantment {B}{G}:
//
//	"Creature tokens you control have "{T}: Add one mana of any
//	 color."
//	 Whenever one or more creature cards leave your graveyard, create
//	 a 0/1 green Plant creature token, then put a +1/+1 counter on
//	 each Plant you control."
//
// The grant is ADR 0093's layer-6 grant; the trigger is Desecrated
// Tomb's batch trigger ("one or more … leave", OncePerBatch). The
// counters go on every Plant you control, the new one included, and
// only after the token has really entered — a creation can pause on a
// CR 616 prompt, so the counters ride the creation's continuation.
//
// No simplification.
const insidiousRootsGrant = "insidious-roots/any-color"

func init() {
	Register(Spec{
		OracleID:     "d75b2b8e-05c9-47da-b359-a867256d78ea",
		Name:         "Insidious Roots",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{AnyColorManaGrant(insidiousRootsGrant)},
		Static:       []game.StaticAbility{GrantAbilities(creatureTokensYouControl, insidiousRootsGrant)},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventZoneMove,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b02CreatureCardLeftYourGraveyard(ev, source, g)
				},
				"Insidious Roots — create a 0/1 Plant, then a +1/+1 counter on each Plant",
				insidiousRootsGrow)),
		},
	})
}

// insidiousRootsGrow is the trigger's resolution.
func insidiousRootsGrow(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	return g.CreateTokensThenForEffect(game.TokenCreation{
		Controller: controller,
		Source:     item.SourceCardID,
		Groups:     []game.TokenGroup{{Template: TokenCard("0/1 green Plant"), Count: 1}},
	}, func(g *game.Game, _ []uuid.UUID) error {
		// asGroupMember (#1463): "each Plant you control" is a set read
		// off the board, not "this permanent".
		ctx := NewContext(g, item).asGroupMember()
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller != controller || !c.HasSubtype("Plant") {
				continue
			}
			if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
