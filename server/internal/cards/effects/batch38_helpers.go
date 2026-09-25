package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch38_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 38 (#401, `edhrec_rank` 3953–4052). Own file per
// the #231 convention; every package-level name carries the b38
// prefix because other batches land beside this one.
//
// What is NOT here, because the catalog already had it: "this
// permanent enters" is b06SelfETB, "each opponent discards a card" is
// eachOpponentDiscardsOne, "N damage to each creature matching" is
// damageEachMatching, "N damage to each opponent" is
// damageToEachOpponent, the tapped-dual land shapes are
// SelfEntersTapped / SelfEntersTappedUnless(
// b40CatchUpDualCondition()) and b36DesertDual, the lord
// builders are TribalAnthem / TribalKeywordGrant, cascade is
// Cascade(), "as this enters, choose a creature type" is
// ChooseCreatureTypeAsEnters plus TribeFilter{Chosen: true},
// convoke is Convoke(), the tutor-to-hand body is b06TutorToHand,
// devotion is devotionTo, and the first-legal-target read is
// b16FirstLegalTargetCard.

// --- damage --------------------------------------------------------

// b38DamageEachPlayer is "… deals N damage to each player" on its own
// — Flame Rift's whole text. EVERY seated player, the source's
// controller included; an eliminated seat is skipped because it is no
// longer a player (CR 800.4).
//
// Not b23DamageEachCreatureAndEachPlayer: that one is the Pyrohemia /
// Pestilence clause, which sweeps the creatures first. A card that
// hits only the players must not touch the board, and the two are
// different printed sentences rather than one with a flag.
//
// The seats are walked in table order and each hit is applied before
// the next is computed, which matters only for the loss-of-life
// triggers that fire in between; nothing here reads a count, so
// there is no snapshot to take.
func b38DamageEachPlayer(ctx *Context, amount int) error {
	if amount <= 0 {
		return nil
	}
	for _, p := range ctx.Game.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if err := (DealDamage{Source: ctx.Source(), Target: p.ID, Amount: amount}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// --- trigger conditions --------------------------------------------

// b38ArtifactPutIntoAnOpponentsGraveyard is Viridian Revel's
// condition: an artifact was put into a graveyard from the
// battlefield, and the graveyard was an OPPONENT's.
//
// Which graveyard a permanent goes to is decided by OWNERSHIP, not
// control (CR 404.3), so the owner is what this reads. A Treasure the
// Revel's controller owns and an opponent controls goes to the
// controller's own graveyard and does not count; an opponent's
// artifact that the Revel's controller stole and then sacrificed goes
// to the opponent's graveyard and does. That is the whole difference
// between this card and Disciple of the Vault, which watches the same
// event over the whole table.
//
// The card is read post-move, so a token is still findable until the
// state-based sweep removes it (CR 111.7).
func b38ArtifactPutIntoAnOpponentsGraveyard(ev game.Event, source *game.Card, g *game.Game) bool {
	if !b23ArtifactPutIntoGraveyardFromBattlefield(ev, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Owner != source.Controller
}

// b38ThisDealtCombatDamageToAnOpponent is Hydra Omnivore's condition:
// the source itself dealt combat damage, and it landed on a PLAYER
// who is an opponent of the source's controller.
//
// Narrower than combatDamageToPlayerBy in two ways the card needs. It
// is this creature, not any creature its controller has — so a second
// Hydra's connection does not fire the first one's ability. And the
// damaged player must be an opponent: a goaded Hydra forced to attack
// its own controller, or one attacking a teammate in a format with
// them, triggers nothing.
func b38ThisDealtCombatDamageToAnOpponent(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	if ev.Source != source.InstanceID {
		return false
	}
	p := g.PlayerByIDForEffect(ev.Target)
	return p != nil && p.ID != source.Controller
}

// --- resolution choices --------------------------------------------

// b38TapOrUntapTarget is Merrow Reejerey's "tap or untap target
// permanent": the target was chosen when the trigger went on the
// stack, and the pick between the two actions is made here, as it
// resolves.
//
// It is a resolution-time option pick rather than a mode, because a
// trigger has no announce-time mode picker and because the printed
// choice really is made on resolution. Either branch is always
// legal — tapping a tapped permanent and untapping an untapped one
// both do nothing, and the card places no condition on the target —
// so the first option is one the chooser can always take, which is
// what the prompt requires.
//
// A target that stopped being legal takes the whole clause with it
// (CR 608.2b) and no question is asked.
func b38TapOrUntapTarget(cardName string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		target := uuid.Nil
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				target = t.ID
				break
			}
		}
		if target == uuid.Nil {
			return nil
		}
		return PickOption{
			Question: cardName + " — tap or untap that permanent",
			Options: []game.ChoiceOption{
				{Label: "Tap it"},
				{Label: "Untap it"},
			},
			Then: func(ctx *Context, index int) error {
				switch index {
				case 0:
					return TapTarget{Target: target}.Apply(ctx)
				case 1:
					return UntapTarget{Target: target}.Apply(ctx)
				default:
					return nil
				}
			},
		}.Apply(ctx)
	}
}

// --- each-player choices -------------------------------------------

// b38EachPlayerReanimatesOne is Exhume's whole text: every seated
// player puts a creature card from THEIR OWN graveyard onto the
// battlefield, each choosing for themselves.
//
// Nothing targets, which is the card. A creature with hexproof or
// protection in a graveyard has neither (those are battlefield
// abilities), but more importantly the caster does not pick, so
// Exhume is a symmetric effect that a deck wins with by having the
// biggest thing in its own yard rather than by choosing.
//
// One prompt per player, addressed to that player and offering only
// their own graveyard, the eachOpponentDiscardsOne posture. A player
// with no creature card in their graveyard is skipped rather than
// prompted — a mandatory instruction with nothing to do does nothing
// (CR 608.2) — and each creature returns under its OWNER's control,
// which for a card in that player's own graveyard is that player.
func b38EachPlayerReanimatesOne(g *game.Game, item *game.StackItem, question string) error {
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.Graveyard == nil {
			continue
		}
		var candidates []uuid.UUID
		for _, c := range p.Graveyard.Cards {
			if c.IsCreature() {
				candidates = append(candidates, c.InstanceID)
			}
		}
		if len(candidates) == 0 {
			continue
		}
		owner := p.ID
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    owner,
			FromPlayer: owner,
			Source:     item.SourceCardID,
			Question:   question,
			Cards:      candidates,
			Min:        1,
			Max:        1,
			// Re-checked on submit: a graveyard can be emptied
			// between the question and the answer.
			Zone: game.ZoneGraveyard,
			Then: func(g *game.Game, picked []uuid.UUID) error {
				for _, id := range picked {
					if err := (ReturnFromGraveyard{
						Target:     id,
						Dest:       game.ZoneBattlefield,
						Controller: owner,
					}).Apply(NewContext(g, item)); err != nil {
						return err
					}
				}
				return nil
			},
		})
	}
	return nil
}

// --- static scopes -------------------------------------------------

// b38SlimedNonHorror is Sludge Monster's static scope: a creature
// with at least one slime counter on it that is not a Horror.
//
// Anyone's creature, including the Sludge Monster's controller's, and
// with no reference to which permanent put the counter there — the
// printed text says "creatures with slime counters on them", so a
// counter from Toxrill, the Corrosive counts too.
//
// Types are the post-layer ones. Layer 4 has settled them before
// either of the statics this scopes (layer 6 and layer 7b) runs, so a
// changeling that something made a Horror is exempt and a Horror that
// lost the type is not.
func b38SlimedNonHorror(target *game.Card, _ *game.Game, _ *game.Card) bool {
	return target.IsCreature() && target.Counters["slime"] > 0 && !target.HasSubtype("Horror")
}

// --- board reads ---------------------------------------------------

// b38FaeriesYouControl counts the Faeries `controller` controls —
// Spell Stutter's scaling tax.
//
// Post-layer subtypes, so a changeling counts and a Faerie that
// something turned into a Wall does not. Creatures only, which is
// what "Faerie" means on a battlefield count: a Faerie ARTIFACT that
// is not a creature would still carry the subtype, but no such card
// exists and the printed clause means the bodies.
func b38FaeriesYouControl(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.HasSubtype("Faerie") {
			n++
		}
	}
	return n
}

// b38OtherCreaturesYouControlNamed counts the creatures `controller`
// controls that are named `name`, excluding the one with instance ID
// `self` — Hare Apparent's "other creatures you control named Hare
// Apparent".
//
// The name is the EFFECTIVE one, so a Clone copying a Hare Apparent
// counts (it is named Hare Apparent) and the card's own 1/1 tokens do
// not (they are named Rabbit). "Other" is by instance rather than by
// name, because every card this clause appears on is legal in
// multiples and the other copies must all count.
func b38OtherCreaturesYouControlNamed(g *game.Game, controller, self uuid.UUID, name string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == self || c.Controller != controller || !c.IsCreature() {
			continue
		}
		if c.Effective().Name == name {
			n++
		}
	}
	return n
}
