package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raggadragga, Goreguts Boss — Legendary Creature — Human Boar {2}{R}{G}, 4/4
// (Commander Legends: Battle for Baldur's Gate):
//
//	"Each creature you control with a mana ability gets +2/+2.
//	 Whenever a creature you control with a mana ability attacks, untap it.
//	 Whenever you cast a spell, if at least seven mana was spent to cast it, untap
//	 target creature. It gets +7/+7 and gains trample until end of turn."
//
// "A mana ability" (CR 605.1) is read from the creature as it is right now,
// printed or granted (game.ManaAbilitiesForCard and ManaTriggersForCard read the
// layered view). The anthem is a layer 7c +2/+2 for the controller's creatures
// that have one; the attack trigger untaps the attacker that triggered it (it does
// not stop the attack: an attacking creature stays attacking, CR 506.4). The cast
// trigger is Prompto Argentum's read of the mana spent on the spell (an
// intervening "if", CR 603.4, asked as it is cast; what was spent never changes),
// with a target creature untapped and given +7/+7 and trample until end of turn.
//
// No simplification.
func init() {
	hasMana := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller && effects.GapB218HasAManaAbility(*target)
	}
	cast := effects.Targeting(
		effects.On(game.EventCast, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
			return effects.YouCast(nil)(ev, source, lki, g) && g.StackItemPaidForEffect(ev.CardID).Spent().Total() >= 7
		}, "Raggadragga, Goreguts Boss — untap target creature; it gets +7/+7 and gains trample until end of turn",
			func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					if err := (effects.BoostUntilEOT{Target: t.ID, Power: 7, Toughness: 7, Label: "Raggadragga — +7/+7"}).Apply(ctx); err != nil {
						return err
					}
					if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Raggadragga — trample"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		effects.TargetCreature("target creature"))
	effects.Register(effects.Spec{
		OracleID:       "a7641ee0-9c0d-4201-9827-a775ed03797c",
		Name:           "Raggadragga, Goreguts Boss",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsYourCasts,
		Static:         []game.StaticAbility{effects.B16Anthem(hasMana, 2, 2)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.AttackDeclaredByYou(ev, source.Controller) || ev.CardID == uuid.Nil {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && effects.GapB218HasAManaAbility(c)
			}, "Raggadragga, Goreguts Boss — untap the attacking creature", effects.OldB345UntapEventCreature),
			cast,
		},
	})
}
