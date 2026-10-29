package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clandestine Chameleon — Creature — Lizard Guest {3}{G}, 4/3:
//
//	"When Clandestine Chameleon enters, you get {TK}{TK}, then you may put
//	 a sticker on a nonland permanent you own.
//	 Clandestine Chameleon has all abilities of ability stickers on other
//	 permanents you own and cards in your graveyard."
//
// The lenders are the ability stickers on the permanents its controller
// owns, other than the Chameleon, and on the cards in its controller's
// graveyard (a sticker stays on a card that goes to a graveyard, CR 123.5).
// "The ability of an ability sticker" is what the sticker grants to the
// object it is on (CR 123.7a), written onto the Chameleon in layer 6 with
// the Chameleon as "this creature" (see HasAllAbilitiesOfAbilityStickers);
// it holds even when the lender has lost its abilities. The Chameleon's own
// stickers give their abilities as any sticker does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "886e00ab-c694-4839-b3fb-7a72048db4b2",
		Name:         "Clandestine Chameleon",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			HasAllAbilitiesOfAbilityStickers(
				"Clandestine Chameleon has all abilities of ability stickers on other permanents you own and cards in your graveyard.",
				selfOnly, clandestineChameleonLenders),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Clandestine Chameleon — you get {TK}{TK}, then you may put a sticker on a nonland permanent you own",
				Do(GetTickets{N: 2}, PutASticker{})),
		},
	})
}

// clandestineChameleonLenders are the stickers on the other permanents the
// source's owner owns and on the cards in their graveyard. It walks the
// zones directly: a layer pass may hold only the read lock.
func clandestineChameleonLenders(g *game.Game, source *game.Card) []game.PlacedSticker {
	var out []game.PlacedSticker
	if g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != source.InstanceID && c.Owner == source.Owner {
				out = append(out, c.Stickers...)
			}
		}
	}
	for _, p := range g.Seats {
		if p == nil || p.ID != source.Controller || p.Graveyard == nil {
			continue
		}
		for i := range p.Graveyard.Cards {
			out = append(out, p.Graveyard.Cards[i].Stickers...)
		}
	}
	return out
}
