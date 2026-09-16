package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch09_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 09 (#302, `edhrec_rank` 1005–1110). Own file per
// the #231 convention; every package-level name carries the b09
// prefix because batch 08 is landing beside this one.
//
// What is NOT here, because main already had it: "an instant or
// sorcery cast by you" is instantOrSorceryCastByYou, "untap up to
// N lands" is untapUpToLands, the reveal-and-tutor body is
// b06TutorToHand, "this permanent enters" is b06SelfETB, the Castle
// condition is b06EntersTappedUnlessLandType, and the mana value of
// a spell on the stack is manaValueOnStack.

// b09OpponentCastInstantOrSorcery is Arasta of the Endless Web's
// condition: an OPPONENT of the source's controller cast an instant
// or sorcery. instantOrSorceryCastByYou with the actor test flipped.
func b09OpponentCastInstantOrSorcery(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && (spell.IsInstant() || spell.IsSorcery())
}

// b09IsHistoric is CR 700.6's "historic": an artifact, a legendary,
// or a Saga — Jhoira, Weatherlight Captain's word. Reads the
// effective characteristic, which for a spell on the stack is the
// printed type line.
func b09IsHistoric(c game.Card) bool {
	if c.IsArtifact() || c.HasSubtype("Saga") {
		return true
	}
	for _, s := range c.Effective().Supertypes {
		if s == "Legendary" {
			return true
		}
	}
	return false
}

// b09SpellCastByYouWithManaValueAtLeast reports whether ev is the
// source's controller casting a spell whose mana value on the stack
// (X included, CR 202.3e) is at least n — Up the Beanstalk's "spell
// with mana value 5 or greater".
func b09SpellCastByYouWithManaValueAtLeast(ev game.Event, source *game.Card, g *game.Game, n int) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	return manaValueOnStack(spell, g.StackItemForEffect(ev.CardID)) >= n
}

// b09IsEquipmentCard is the SearchLibrary predicate for "an Equipment
// card" — Steelshaper's Gift, Stoneforge Mystic. Subtype, not name:
// a card in a library has no layer cache, so this is the printed
// type line.
func b09IsEquipmentCard(c game.Card) bool { return c.HasSubtype("Equipment") }

// b09IsCheapInstantOrSorceryCard is Spellseeker's "an instant or
// sorcery card with mana value 2 or less". Off the stack, so X is
// zero (CR 202.3e) and a Fireball is a legal find.
func b09IsCheapInstantOrSorceryCard(c game.Card) bool {
	return (c.IsInstant() || c.IsSorcery()) && manaValueOf(c) <= 2
}

// b09CounterThenUntapLands is the shared OnResolve of Rewind and
// Unwind: counter the targeted spell, then untap up to n lands. The
// untap runs even when the target has already left the stack —
// CR 608.2b would fizzle the whole spell in that case, and the
// engine already does so before OnResolve fires, so the only way
// to get here with no target is a spell countered by something the
// target predicate does not check, in which case "untap" is still
// the printed instruction.
//
// "Untap up to n lands" is the Snap posture: the printed clause is
// a resolution-time choice with no "you control", and no
// pick-a-permanent prompt exists for a spell, so it untaps the
// first n tapped lands the caster controls in battlefield order —
// never an opponent's, never stronger than printed, and declared on
// both cards.
func b09CounterThenUntapLands(n int) func(item *game.StackItem, ctx *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard {
			if err := (CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
		}
		return untapUpToLands(ctx.Game, ctx.Controller(), n)
	}
}

// b09OpponentsHoldingCards counts the opponents of `controller` who
// have at least one card in hand — Syphon Mind's draw count. Each of
// them discards exactly one card, so the number of cards that will
// be discarded is fixed the moment the spell resolves.
func b09OpponentsHoldingCards(ctx *Context) int {
	n := 0
	for _, opp := range ctx.Opponents() {
		if p := ctx.PlayerByID(opp); p != nil && p.Hand.Size() > 0 {
			n++
		}
	}
	return n
}

// b09ArchonOfCrueltyTrigger is the body of Archon of Cruelty's
// enters-or-attacks trigger, read off the item's target slot:
//
//	"target opponent sacrifices a creature or planeswalker of their
//	 choice, discards a card, and loses 3 life. You draw a card and
//	 gain 3 life."
//
// The sacrifice is the opponent's own choice — PlayerSacrificesForEffect
// queues them a prompt over their own creatures and planeswalkers,
// which is what "of their choice" means and why it is not a target
// (hexproof is irrelevant to it). The discard is their choice too,
// through the same pending-discard modal Mind Rot uses. The life
// loss, the draw and the life gain run at once; the two prompts
// settle whenever the opponent answers. Package-level so the
// triggered item captures nothing.
func b09ArchonOfCrueltyTrigger(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
		return nil
	}
	victim := item.Targets[0].ID
	ctx := NewContext(g, item)
	if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
		return nil
	}
	g.PlayerSacrificesForEffect(item.SourceCardID, victim,
		sacrificeSpec("a creature or planeswalker", Or(Creature(), Planeswalker())),
		"Archon of Cruelty — sacrifice a creature or planeswalker")
	g.DiscardChoiceForEffect(victim, 1)
	if err := g.ChangePlayerLifeForEffect(item.SourceCardID, victim, -3); err != nil {
		return err
	}
	if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return GainLife{Player: item.Controller, Amount: 3}.Apply(ctx)
}

// b09SourceStillOnBattlefield is the guard every "put a counter on
// this creature" trigger needs: the source may have left between
// the trigger going on the stack and resolving, and AddCounter does
// not gate on zone.
func b09SourceStillOnBattlefield(g *game.Game, item *game.StackItem) bool {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	return z != nil && z.Kind == game.ZoneBattlefield
}
