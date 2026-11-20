package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dogmeat, Ever Loyal — Legendary Creature — Dog {R}{G}{W}, 3/3 (Fallout,
// #5948):
//
//	"When Dogmeat enters, mill five cards, then return an Aura or Equipment
//	 card from your graveyard to your hand.
//	 Whenever a creature you control that's enchanted or equipped attacks,
//	 create a Junk token. (It's an artifact with "{T}, Sacrifice this token:
//	 Exile the top card of your library. You may play that card this turn.
//	 Activate only as a sorcery.")"
//
// The pick is asked after the mill (Lost Monarch of Ifnir's shape), over the
// Aura and Equipment cards in the graveyard then, so a card the mill just
// put there can be returned; it is not optional while there is one. The
// second ability is one trigger per attacking creature you control that has
// an Aura or an Equipment attached as it attacks (an Aura that is yours or
// not, "enchanted" being any Aura); the token is the printed Junk.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5232fcc0-5dba-4569-8e99-c0b2c230e9f0",
		Name:         "Dogmeat, Ever Loyal",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dogmeat — mill five cards, then return an Aura or Equipment card from your graveyard to your hand", dogmeatEnters),
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !attackDeclaredByYou(ev, source.Controller) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && (b18AurasAttachedTo(g, c.InstanceID) > 0 || isEquippedCreature(g, c))
			}, "Dogmeat — create a Junk token", Do(CreateToken{Template: JunkToken(), N: 1})),
		},
	})
}

func dogmeatEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (MillCards{Player: item.Controller, N: 5}).Apply(ctx); err != nil {
		return err
	}
	candidates := graveyardCardIDs(ctx, item.Controller, func(c game.Card) bool {
		return c.HasSubtype("Aura") || c.HasSubtype("Equipment")
	})
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Dogmeat — return an Aura or Equipment card from your graveyard to your hand",
		Cards:      candidates,
		Min:        1,
		Max:        1,
		ThenKey:    dogmeatReturn,
		LiveZone:   game.ZoneGraveyard,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// dogmeatReturn is the chosen card going to its owner's hand. An on-disk
// key: never renamed, never reused.
var dogmeatReturn = askedPick("choose-cards/dogmeat-return-aura-or-equipment", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
