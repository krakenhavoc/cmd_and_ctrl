package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Light-Paws, Emperor's Voice — Legendary Creature — Fox Advisor
// {1}{W}, 2/2 (EDHREC rank 3097):
//
//	"Whenever an Aura you control enters, if you cast it, you may
//	 search your library for an Aura card with mana value less than
//	 or equal to that Aura and with a different name than each Aura
//	 you control, put that card onto the battlefield attached to
//	 Light-Paws, then shuffle."
//
// The Aura voltron commander. The trigger is an Aura entering under
// the controller's control that was CAST — its move onto the
// battlefield came from the stack (b16EnteredFromStack), so a
// fetched or reanimated Aura, including the one this very ability
// finds, does not chain. The intervening "if you cast it" cannot
// change between trigger and resolution, so it is checked once.
// The "you may" is the trigger's optional prompt; the search
// admits Aura cards with mana value at most the entering Aura's
// (read at trigger time) and a name no Aura the controller controls
// has (read at resolution, as printed), and the found card is put
// onto the battlefield and then attached to Light-Paws in the
// search's continuation (b29SearchAuraAttachedToSource). Nothing is
// searched if Light-Paws has left the battlefield: there is nothing
// to attach the card to.
//
// Sandbox simplification, declared, weaker than printed: an Aura
// found this way that cannot legally enchant Light-Paws ("enchant
// land", "enchant creature you don't control") is attached anyway
// and then put into the graveyard by the CR 704.5n state-based
// action, where printed it would stay in the library. The search
// offers every Aura card, because the enchant clause of a card in
// the library is not readable from a library predicate.
func init() {
	Register(Spec{
		OracleID:     "1718a442-b878-4690-b608-a013de3d79fc",
		Name:         "Light-Paws, Emperor's Voice",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"An Aura found by the search that can't legally enchant Light-Paws goes to your graveyard instead of staying in your library."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := b29AuraYouCastEntered(ev, source, g)
				return ok
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Light-Paws, Emperor's Voice — search for an Aura to attach to it?"},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				maxMV := 0
				if aura, ok := g.LookupCardForEffect(ev.CardID); ok {
					maxMV = aura.ManaValue()
				}
				return game.NewTriggeredItem(source, b29LightPawsLabel,
					func(g *game.Game, item *game.StackItem) error {
						return b29SearchAuraAttachedToSource(g, item, maxMV)
					})
			},
		}},
	})
}
