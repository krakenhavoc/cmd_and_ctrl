package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Evelyn, the Covetous — Legendary Creature — Vampire Rogue
// {2}{U/B}{B}{B/R}, 2/5 (Streets of New Capenna, #5876):
//
//	"Flash
//	 Whenever Evelyn or another Vampire you control enters, exile the top
//	 card of each player's library with a collection counter on it.
//	 Once each turn, you may play a card from exile with a collection
//	 counter on it if it was exiled by an ability you controlled, and you
//	 may spend mana as though it were mana of any color to cast it."
//
// The trigger is "this or another <type> you control enters" over Vampire
// (a word of the card's text, CR 612.2). It exiles the top card of each
// library one at a time in APNAP order (each_player_in_turn.go), and once
// every library is done each card that reached exile gets its collection
// counter, an ordinary counter on the exiled card (CR 122.1) that MoveCard
// clears when the card leaves exile (Tinybones' stash counter).
//
// The play is a STANDING permission (ADR 0066) read off the battlefield
// on every query, so it ends when this Evelyn leaves:
//
//   - "with a collection counter on it" is WithCounter;
//   - "if it was exiled by an ability you controlled" is
//     ExiledByHolderAbility (the #5876 seam): the exile stamps the
//     controller of the resolving ability on the card
//     (Card.ExiledByAbilityOf), so a card another Evelyn of yours exiled
//     qualifies, and one an opponent's Evelyn exiled does not, whoever
//     owns it;
//   - "once each turn" is Serra Paragon's one-use budget, PerType
//     PermissionTypePermanent, which every play or cast spends (Johann,
//     Apprentice Sorcerer casts instants with it). It is kept per Evelyn
//     OBJECT on the turn tally, so two Evelyns give two plays, and a new
//     Evelyn is a new object with a fresh use (CR 400.7);
//   - "spend mana as though it were mana of any color" is AnyColor
//     (CR 609.4b), which leaves a {C} requirement standing;
//   - "play" covers a land, which still uses a land drop (CR 305.2), and
//     the card's own timing is unchanged.
//
// No simplification.
const evelynCollectionCounter = "collection"

func init() {
	Register(Spec{
		OracleID:        "4ab6f615-1f04-4ed2-88b2-49faa8356dd0",
		Name:            "Evelyn, the Covetous",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, evelynOrAnotherVampireEntered,
				"Evelyn, the Covetous — exile the top card of each player's library with a collection counter on it",
				evelynExileTopOfEachLibrary),
		},
		CastPermissions: []game.CastPermission{{
			Zone: game.ZoneExile,
			Filter: game.PermissionFilter{
				WithCounter:           evelynCollectionCounter,
				ExiledByHolderAbility: true,
			},
			AnyColor: true,
			PerType:  []string{game.PermissionTypePermanent},
			Label:    "Once each turn, play a card with a collection counter exiled by an ability you controlled (Evelyn, the Covetous)",
		}},
	})
}

// evelynOrAnotherVampireEntered is "whenever Evelyn or another Vampire you
// control enters".
func evelynOrAnotherVampireEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	if !ok {
		return false
	}
	return c.InstanceID == source.InstanceID || c.HasSubtype(source.TextSubtype("Vampire"))
}

// evelynExileTopOfEachLibrary exiles the top card of each player's
// library, then puts a collection counter on each card that landed.
func evelynExileTopOfEachLibrary(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return eachPlayerInTurn(g, ctx.ItemTail(evelynExileStep), apnapPlayers(g), nil,
		ctx.ItemTail(evelynCollect), evelynExileAct)
}

// evelynExileStep is one library of the chain. An on-disk key.
var evelynExileStep = registerEachPlayerStep("evelyn-the-covetous/exile-step", evelynExileAct)

func evelynExileAct(ctx *Context, player uuid.UUID, _ int, step game.KeyedTail) error {
	return ctx.Game.ExileTopKeyedForEffect(player, ctx.Controller(), 1, step)
}

// evelynCollect puts the collection counter on each exiled card still in
// exile. An on-disk key.
var evelynCollect = CardTail("evelyn-the-covetous/collect", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		z := ctx.Game.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if err := ctx.Game.AddCounterForEffect(id, evelynCollectionCounter, 1); err != nil {
			return err
		}
	}
	return nil
})
