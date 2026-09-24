package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch16_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 16 (#309, `edhrec_rank` 1727–1829). Own file per
// the #231 convention; every package-level name carries the b16
// prefix because batch 15 landed beside this one.
//
// What is NOT here, because main already had it: devotion is
// devotionTo, "this permanent enters" is b06SelfETB, "another
// creature you control enters" is b13AnotherCreatureYouControlEntered,
// "whenever you cast an instant or sorcery" is
// instantOrSorceryCastByYou, "a permanent spell resolving" is
// nothing at all (the event log knows — b16EnteredFromStack), a
// tutor to hand is b06TutorToHand, the basic-land fetch body is
// b07SearchBasicOntoBattlefield, the per-label "one or more" dedup is
// OncePerBatch, the once-per-turn tally is
// b11TriggeredThisTurn, the resolution tally is b15ResolvedThisTurn,
// a dead creature's power is b13LastKnownPower, the attackers a
// player controls are b13AttackingCreaturesYouControl, and "any
// number of lands you control, untapped" is b02UntapLandsYouControl.

// --- token templates ---------------------------------------------

// b16LanderToken is the Edge of Eternities Lander: a colorless
// artifact with "{2}, {T}, Sacrifice this token: Search your library
// for a basic land card, put it onto the battlefield tapped, then
// shuffle." The ability rides the template the way Food's and Clue's
// do, because a token has no oracle ID for the catalog to key one
// on. The search is the S22 chooser; a Lander cracked with no basic
// left still shuffles.
func b16LanderToken() game.Card { return tokenFromCatalog(printedB16LanderToken) }

// printedB16LanderToken is the Lander as PRINTED —
// the ability included. It is the catalog's entry for this token
// (token_catalog.go, #521): the ability is registered from here at
// boot, and the template that reaches the battlefield carries the
// key that finds it rather than the closure itself.
func printedB16LanderToken() tokenTemplate {
	return tokenTemplate{
		Slug: "lander",
		Card: game.Card{
			Name:     "Lander",
			TypeLine: "Token Artifact — Lander",
		},
		Text: "{2}, {T}, Sacrifice this token: Search your library for a basic land card, " +
			"put it onto the battlefield tapped, then shuffle.",
		Activated: []game.ActivatedAbilityShape{{
			Label: "{2}, {T}, Sacrifice this token: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle",
			Cost: game.AbilityCost{
				Tap:           true,
				SacrificeSelf: true,
				Mana:          "{2}",
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b07SearchBasicOntoBattlefield(g, item, item.Controller, true, false,
					"Lander — a basic land, onto the battlefield tapped")
			},
		}},
	}
}

// --- statics -----------------------------------------------------

// b16Anthem is a Layer 7c "+P/+T" static over whatever `applies`
// selects — Death Baron's Skeletons and Zombies, Lyra's Angels.
func b16Anthem(applies func(target *game.Card, g *game.Game, source *game.Card) bool, power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: applies,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += power
			c.Toughness += toughness
		},
	}
}

// b16GrantKeywords is a Layer 6 keyword grant over whatever
// `applies` selects — "Creatures you control have double strike and
// lifelink" (True Conviction), "… have haste" (Urabrask), "… have
// vigilance" (Brave the Sands). Keywords must be canonical tokens the
// engine honours; deduped on apply so a creature that prints the
// keyword gets one badge.
func b16GrantKeywords(applies func(target *game.Card, g *game.Game, source *game.Card) bool, keywords ...string) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		AppliesTo: applies,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			for _, kw := range keywords {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, kw)
			}
		},
	}
}

// b16CreaturesYouControl is the "creatures you control" AppliesTo,
// the source included.
func b16CreaturesYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.Controller == source.Controller
}

// b16OtherCreaturesYouControlOfSubtype is "other <subtype>s you
// control" — Lyra's Angels. Effective subtypes, so a changeling
// counts.
func b16OtherCreaturesYouControlOfSubtype(subtype string) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID != source.InstanceID && target.IsCreature() &&
			target.Controller == source.Controller && target.HasSubtype(subtype)
	}
}

// b16DeathBaronApplies is "Skeletons you control and other Zombies
// you control" — a Skeleton Zombie Baron would count itself, as
// printed, and a plain Zombie Baron does not.
func b16DeathBaronApplies(target *game.Card, _ *game.Game, source *game.Card) bool {
	if !target.IsCreature() || target.Controller != source.Controller {
		return false
	}
	if target.HasSubtype("Skeleton") {
		return true
	}
	return target.InstanceID != source.InstanceID && target.HasSubtype("Zombie")
}

// --- trigger conditions ------------------------------------------

// b16EnteredFromStack is "if you cast it" for an entering permanent
// (Zacama). Every battlefield entry emits an EventZoneMove
// immediately before its EventETB, and the only zone a permanent
// CARD arrives from having been cast is the stack — a reanimated,
// fetched, flickered or played-as-a-land permanent comes from
// somewhere else. Read back off the log: the most recent zone move
// onto the battlefield for this card.
func b16EnteredFromStack(g *game.Game, cardID uuid.UUID) bool {
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Kind == game.EventZoneMove && ev.CardID == cardID && ev.NewZone == game.ZoneBattlefield {
			return ev.OldZone == game.ZoneStack
		}
	}
	return false
}

// b16CardLeftYourGraveyard is "whenever one or more cards leave your
// graveyard" (Teval's Judgment), one event at a time: a move out of
// a graveyard for a card the source's controller OWNS — "your
// graveyard" is the one you own. Two event kinds carry one: an
// EventZoneMove (exile, a regrowth, a reanimation, a graveyard
// sweep) and an EventCast whose OldZone is the graveyard (a
// flashback — the cast path emits the cast with the zone-move
// fields rather than a second event). The card is read post-move,
// wherever it went. The "one or more" batching is the caller's
// dedup.
func b16CardLeftYourGraveyard(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventZoneMove && ev.Kind != game.EventCast {
		return false
	}
	if ev.OldZone != game.ZoneGraveyard || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Owner == source.Controller
}

// b16SelfOrAnotherCreatureYouControlDied is Vengeful Bloodwitch's
// condition: the source itself died, or a creature its controller
// controlled did. The source's own death arrives with the card
// already in the graveyard, which cardDied reads by ID.
func b16SelfOrAnotherCreatureYouControlDied(ev game.Event, source *game.Card, g *game.Game) bool {
	if cardDied(ev, source) {
		return true
	}
	dead, ok := diedCreature(ev, g)
	return ok && dead.Controller == source.Controller
}

// b16AnotherCreatureYouControlEnteredOrDied is Daxos's condition on
// one ability watching two kinds.
func b16AnotherCreatureYouControlEnteredOrDied(ev game.Event, source *game.Card, g *game.Game) bool {
	if b13AnotherCreatureYouControlEntered(ev, source, g) {
		return true
	}
	if ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := diedCreature(ev, g)
	return ok && dead.Controller == source.Controller
}

// b16PlayerAttackedWithAtLeast is Aurelia's condition: an attacker
// was just declared, and its controller now has at least n creatures
// attacking. "Whenever a player attacks with N or more creatures" is
// ONE trigger per declaration, and the engine emits EventAttack per
// creature, so the ability fires on the declaration that reaches n
// and declines every later one — every later event of the SAME batch
// (OncePerBatch, keyed on Event.Batch; see AGENTS.md §7) and,
// because the sandbox lets attackers be declared after the trigger
// has resolved, the rest of the turn (b11TriggeredThisTurn). Only
// the active player attacks in a turn and the engine has no extra
// combats, so
// once-per-turn is once-per-declaration. Weaker than printed under
// an extra combat, never stronger.
func b16PlayerAttackedWithAtLeast(ev game.Event, source *game.Card, g *game.Game, n int, label string) bool {
	if ev.Kind != game.EventAttack || ev.Actor == uuid.Nil {
		return false
	}
	if len(b13AttackingCreaturesYouControl(g, ev.Actor)) < n {
		return false
	}
	return !b11TriggeredThisTurn(g, source.InstanceID, label)
}

// b16YouAttackedAPlayer is Horizon Explorer's condition: a creature
// the source's controller controls was declared attacking a PLAYER,
// deduplicated to one trigger per declaration the Adeline way. The
// engine has no attack-a-planeswalker path, so every declaration is
// at a player; the check is kept so the day one exists this stays
// honest.
func b16YouAttackedAPlayer(ev game.Event, source *game.Card, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	return g.PlayerByIDForEffect(ev.Target) != nil
}

// b16CreatedATokenThisTurn reports whether `player` created a token
// this turn — Bennie Bracks's intervening if. EventTokenCreated
// carries the creator in Actor, and the per-turn tally counts it
// there.
func b16CreatedATokenThisTurn(g *game.Game, player uuid.UUID) bool {
	return g.TurnTallyFor(player).TokensCreated > 0
}

// --- board reads -------------------------------------------------

// b16CardTypesInGraveyard counts the distinct card types among the
// cards in `player`'s graveyard — delirium (Demonic Counsel). Printed
// type lines: a card in a graveyard has no layer cache and nothing
// changes its types there.
func b16CardTypesInGraveyard(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	seen := map[string]bool{}
	for _, c := range p.Graveyard.Cards {
		_, types, _ := game.ParseTypeLine(c.TypeLine)
		for _, t := range types {
			seen[t] = true
		}
	}
	return len(seen)
}

// b16ColorlessCreaturesControlled counts the colorless creatures
// `player` controls — Tomb of the Spirit Dragon. Effective colours,
// so a creature painted by a colour-setting effect stops counting.
func b16ColorlessCreaturesControlled(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() && c.IsColorless() {
			n++
		}
	}
	return n
}

// --- effect bodies -----------------------------------------------

// b16UntapAllYouControlMatching untaps every tapped permanent
// `controller` controls that passes `match` — Zacama's lands,
// Unstoppable Plan's nonland permanents. Snapshot then untap, so the
// untap events don't disturb the walk.
func b16UntapAllYouControlMatching(ctx *Context, controller uuid.UUID, match func(game.Card) bool) error {
	var ids []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.Tapped && match(c) {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b16DrawThenDiscard is "draw N cards, then discard M" in the printed
// order (Relic of Sauron's two-for-one): the draws land first so a
// drawn card is a legal discard, and the discard is the player's
// choice, queued as a prompt.
func b16DrawThenDiscard(g *game.Game, item *game.StackItem, draw, discard int) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: item.Controller, N: draw}).Apply(ctx); err != nil {
		return err
	}
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player: item.Controller,
		Source: item.SourceCardID,
		N:      discard,
	})
	return nil
}

// b16FirstLegalTargetCard returns the first announce-time target
// slot that is still a legal battlefield card, for a single-target
// ability whose target may have left in response.
func b16FirstLegalTargetCard(ctx *Context) (uuid.UUID, bool) {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return t.ID, true
		}
	}
	return uuid.Nil, false
}

// --- replacements ------------------------------------------------

// b16LandsYouControlEnterUntapped is Horizon Explorer's "Lands you
// control enter untapped": a CR 614 replacement that clears the
// enters-tapped flag on a land entering under the source's
// controller's control.
//
// It applies only once the flag is SET — by the land's own
// enters-tapped replacement, by an opponent's Kismet, or by the
// fetching effect's "onto the battlefield tapped" clause, which the
// search path seeds onto the event. Gating on the flag is what makes
// the CR 616 order fall out without a prompt: with the flag clear
// this effect is not yet applicable, the tapping effect applies
// alone, and on the next iteration of the apply-loop this one is the
// only effect left. The printed card gets the same answer from the
// affected player choosing the order, and no player would choose the
// other way.
//
// There is one case the flag gate does NOT make silent: a land
// FETCHED tapped that also carries its own enters-tapped clause. The
// search seeds the flag before the window opens, so both effects are
// applicable on the first gather and CR 616.1 really does ask the
// controller to order them — which is the printed interaction, and
// the answer is the player's.
//
// That case used to be a declared retreat here (a fetched Guildgate
// was left alone), because the entry site could not pause: the prompt
// was queued, nobody could answer it usefully and the land was
// stranded in the library. #478 closed that — the search entry is
// resumable and carries its tail across the pause — so the retreat
// was describing a hole that had already been filled, and #732 took
// it out. Every land is as printed now.
func b16LandsYouControlEnterUntapped() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || !ev.EntersTapped {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			if !ok || !entering.IsLand() || entering.Controller != src.Controller {
				return false
			}
			return true
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.EntersTapped = false
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: "Horizon Explorer: enters untapped",
	}
}
