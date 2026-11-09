package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// summonersGrimoireGrant is the attack trigger Summoner's Grimoire gives
// its equipped creature.
const summonersGrimoireGrant = "summoners-grimoire/put-creature-from-hand"

// Summoner's Grimoire — Artifact — Book Equipment {3}{G} (Final Fantasy,
// #4462):
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 Equipped creature is a Shaman in addition to its other types and has
//	 "Whenever this creature attacks, you may put a creature card from
//	 your hand onto the battlefield. If that card is an enchantment card,
//	 it enters tapped and attacking."
//	 Abraxas — Equip {3}"
//
// Job select, the Shaman type addition and equip are the set's shared
// pieces (Monk's Fist). The attack trigger is the equipped creature's own
// (ADR 0093, Conformer Shuriken's shape): "you" is its controller. Its put
// is Champion of Rhonas's: the card is put, not cast, so it triggers no
// cast trigger, and the "may" is a real decline. A creature card that is
// also an enchantment card (an enchantment creature) enters tapped and
// attacking (CR 506.3c, CR 508.4): its controller says what it attacks
// (any player, planeswalker or battle they could attack, Interceptor's
// question with the hand as the zone it must still be in), and since it
// was never declared as an attacker it triggers no "whenever attacks".
// Any other creature card enters untapped and not attacking. Abraxas is an
// ability word.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d4453b3e-733e-4038-81da-135a6a03079c",
		Name:         "Summoner's Grimoire",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: summonersGrimoireGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisAttacks("Summoner's Grimoire — you may put a creature card from your hand onto the battlefield; if it's an enchantment card, it enters tapped and attacking",
					summonersGrimoireAttack),
			},
			Text: "Whenever this creature attacks, you may put a creature card from your hand onto the battlefield. If that card is an enchantment card, it enters tapped and attacking.",
		}},
		Static: []game.StaticAbility{
			AttachedIsAlso("Shaman"),
			GrantAbilitiesToAttached(summonersGrimoireGrant),
		},
		Triggered: []game.TriggeredAbility{JobSelect("Summoner's Grimoire")},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
	// "Tapped and attacking" from a hand: the form's own keyed answer to
	// "what does it attack?" (put_attacking.go).
	putAttackingThens[putAttackingForm{from: game.ZoneHand}] = putAttackingFromHand
}

// putAttackingFromHand is the keyed answer to "what does it attack?" for a
// card put from the hand (Summoner's Grimoire).
var putAttackingFromHand = OptionPickThen("option-pick/put-attacking-from-hand",
	putAttackingAnswer(putAttackingForm{from: game.ZoneHand}))

// summonersGrimoireAttack asks the equipped creature's controller for a
// creature card in their hand (a "may": the floor is zero).
func summonersGrimoireAttack(g *game.Game, item *game.StackItem) error {
	cards := handCardsMatching(g, item.Controller, Creature())
	if len(cards) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Destination: game.ChooseOntoBattlefield,
		Chooser:     item.Controller,
		Source:      item.SourceCardID,
		Question:    "Summoner's Grimoire — you may put a creature card from your hand onto the battlefield",
		Cards:       cards,
		Min:         0,
		Max:         1,
		ThenKey:     summonersGrimoirePut,
		LiveZone:    game.ZoneHand,
		Asked:       game.NewAskedItem(item),
	})
	return nil
}

// summonersGrimoirePut puts the creature card picked: tapped and
// attacking when it is an enchantment card, otherwise as it is. An
// on-disk key.
var summonersGrimoirePut = askedPick("choose-cards/summoners-grimoire-put", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	id := r.Picked[0]
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	if c.HasCardType("enchantment") {
		return putAttackingYourChoice(ctx, id, game.ZoneHand, "Summoner's Grimoire")
	}
	return ctx.Game.PutFromHandOntoBattlefieldKeyedForEffect(id, game.HandEntryOptions{Controller: r.Chooser}, game.KeyedTail{})
})
