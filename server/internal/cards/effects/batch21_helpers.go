package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch21_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 21 (#383, `edhrec_rank` 2235–2334), the first
// slice of the second 2000. Own file per the #231 convention; every
// package-level name carries the b21 prefix because other batches
// land beside this one.
//
// What is NOT here, because main already had it: "this permanent
// enters" is b06SelfETB, "a creature you control entered" is
// enteredUnderYourControl, "another creature you control enters"
// is b13AnotherCreatureYouControlEntered, "an opponent's creature
// died" is b18OpponentsCreatureDied, "an opponent discarded" is
// b18OpponentDiscarded, the enchantress condition is
// enchantmentSpellCastByYou, the Elf-spell condition is
// b17ElfSpellCastByYou, the once-per-turn tally is
// b11TriggeredThisTurn, "each opponent loses N" is
// eachOpponentLosesLife, "each opponent loses 1 and you gain 1" is
// drainEachOpponent, the tapped-creature predicate is
// tappedPermanent, the fetch body is fetchBasicTapped, the
// destroy-the-pick body is destroyChosenTargetTrigger, the DMU
// duals are rows in dominaria_united_duals.go, and the Goblin /
// Treasure / Food / Elf Warrior templates live in tokens.go and
// batch13_helpers.go.

// --- token templates ---------------------------------------------

// b21TappedAttackingGoblin is General Kreat's 1/1 red Goblin, stamped
// tapped so CreateTokensAttackingForEffect — which copies the
// template and sets only the attack — puts it in tapped and
// attacking.
func b21TappedAttackingGoblin() game.Card {
	tmpl := RedGoblinToken()
	tmpl.Tapped = true
	return tmpl
}

// --- trigger conditions ------------------------------------------

// b21SelfEnteredOrAttacked is "whenever this creature enters or
// attacks" (Hazel's Brewmaster) — one printed ability with two
// trigger conditions, watching EventETB and EventAttack on one
// declaration. Both kinds carry the card in CardID (Sun Titan's
// shape).
func b21SelfEnteredOrAttacked(ev game.Event, source *game.Card) bool {
	return (ev.Kind == game.EventETB || ev.Kind == game.EventAttack) && ev.CardID == source.InstanceID
}

// b21CreatureOfSubtypeYouControlAttacked is "whenever a <Subtype> you
// control attacks" — a Vampire for Sanctum Seeker, a Goblin for
// General Kreat. The attacker is read live (it is on the battlefield
// as it is declared), so effective subtypes apply and a changeling
// counts. The source itself qualifies when it carries the subtype:
// the printed text is "a Vampire you control", not "another".
func b21CreatureOfSubtypeYouControlAttacked(ev game.Event, source *game.Card, g *game.Game, subtype string) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.HasSubtype(subtype)
}

// b21AngelYouControlEntered is Seraph Sanctuary's second trigger:
// an Angel entered under the controller's control. Effective
// subtypes, so a changeling counts; the Sanctuary itself is a land
// and can never match, so no "another" guard is needed.
func b21AngelYouControlEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsCreature() && c.HasSubtype("Angel")
}

// b21DragonYouControlDealtDamage is Wrathful Red Dragon's condition:
// a Dragon the controller controls was dealt damage — combat or not,
// by anyone's source, the Dragon itself included. EventDealDamage
// names the damaged permanent in Target; damage is marked before the
// state-based sweep runs, so a Dragon that took lethal is still on
// the battlefield when its event fires and is read live.
func b21DragonYouControlDealtDamage(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Target == uuid.Nil {
		return false
	}
	if z := g.FindCardZoneForEffect(ev.Target); z == nil || z.Kind != game.ZoneBattlefield {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	return ok && c.Controller == source.Controller && c.HasSubtype("Dragon")
}

// b21ArtifactOrCreatureYouControlDied is the granted trigger on Agent
// of the Iron Throne: an artifact or creature the controller
// controlled was put into a graveyard from the battlefield. The dead
// card is read post-move — its printed type line and its controller
// survive the move — so a Treasure, a Clue, a token creature and a
// nontoken one all count, as printed.
func b21ArtifactOrCreatureYouControlDied(ev game.Event, source *game.Card, g *game.Game) (game.Card, bool) {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return game.Card{}, false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || c.Controller != source.Controller {
		return game.Card{}, false
	}
	if !c.IsArtifact() && !c.IsCreature() {
		return game.Card{}, false
	}
	return c, true
}

// --- board reads -------------------------------------------------

// b21ControlsCommanderCreatureYouOwn reports whether `player`
// controls a creature that is a commander THEY OWN — the permanent a
// Background's "Commander creatures you own have …" hangs its
// granted ability on. Owner matters: a stolen opponent's commander is
// not yours, and your own commander under an opponent's control is
// not something you control.
func b21ControlsCommanderCreatureYouOwn(g *game.Game, player uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCommander && c.Owner == player && c.Controller == player && c.IsCreature() {
			return true
		}
	}
	return false
}

// b21AnyGraveyardHasACard reports whether any seated player's
// graveyard holds at least one card — the legal-target set of
// "target card from a graveyard" is non-empty. Hazel's Brewmaster
// splits its trigger on this answer (see the card file).
func b21AnyGraveyardHasACard(g *game.Game) bool {
	for _, p := range g.Seats {
		if p != nil && p.Graveyard != nil && p.Graveyard.Size() > 0 {
			return true
		}
	}
	return false
}

// --- target specs ------------------------------------------------

// b21TargetAnyNonDragon is "any target that isn't a Dragon" —
// Wrathful Red Dragon's reflected damage. TargetAny's set (a player,
// a creature, a planeswalker, a battle) minus permanents with the
// Dragon subtype; players are never Dragons, so the player half is
// untouched.
func b21TargetAnyNonDragon() *game.TargetSpec {
	spec := TargetAny()
	spec.Label = "any target that isn't a Dragon"
	base := spec.CardOK
	spec.CardOK = func(g *game.Game, caster uuid.UUID, c game.Card, z game.ZoneKind) bool {
		return base(g, caster, c, z) && !c.HasSubtype("Dragon")
	}
	return spec
}

// --- effect bodies -----------------------------------------------

// b21DrainEachOpponentAndGainTheTotal is "each opponent loses N life.
// You gain life equal to the life lost this way" — Kokusho at a fixed
// 5, Gray Merchant at your devotion to black, Exsanguinate at X, Debt
// to the Deathless at 2X. Life loss, not damage, so no prevention
// shield and no damage doubler sees it.
//
// "LIFE LOST THIS WAY" IS NOT N × OPPONENTS. It is what each opponent
// really lost, which is the same number only while nothing is replacing
// anybody's life loss. #793: the gain therefore rides
// LoseLifeEachThenForEffect's continuation, which reports the true
// total once every opponent's own CR 614 window has settled — including
// the ones that paused on a CR 616 ordering prompt, where reading the
// life totals back on the next line saw nothing at all and gained zero.
func b21DrainEachOpponentAndGainTheTotal(g *game.Game, item *game.StackItem, n int) error {
	if n <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
	controller, source := item.Controller, ctx.Source()
	return g.LoseLifeEachThenForEffect(source, ctx.Opponents(), n, func(g *game.Game, totalLost int) error {
		if totalLost <= 0 {
			return nil
		}
		return g.ChangePlayerLifeForEffect(source, controller, totalLost)
	})
}

// b21DamageEachOpponentAndTheirCreaturesAndWalkers is End the
// Festivities' body: N damage to each opponent, and to each creature
// and planeswalker an opponent controls. The permanents are read
// once before any damage is dealt (MatchingBattlefield snapshots),
// so nothing is hit twice and a creature that dies later does so at
// the state-based sweep, as printed.
func b21DamageEachOpponentAndTheirCreaturesAndWalkers(ctx *Context, n int) error {
	for _, opp := range ctx.Opponents() {
		if err := (DealDamage{Source: ctx.Source(), Target: opp, Amount: n}).Apply(ctx); err != nil {
			return err
		}
	}
	return damageEachMatching(ctx, And(OpponentControls(), Or(Creature(), Planeswalker())), n)
}

// b21RevealUntilBasicLandToHand is Hermit Druid's body: reveal cards
// from the top of `player`'s library until a basic land card is
// revealed; that card goes to hand and every other revealed card to
// the graveyard.
//
// The library is walked read-only first to find the first basic land
// from the top, then the cards above it are milled (one EventMill
// each, so a mill payoff sees them) and the land moves from the
// library straight to hand through the bounce-to-hand path, which
// finds a card in any zone and emits one ordinary zone move. The land
// never touches the graveyard and is never searched for or drawn —
// no mill, search or draw trigger sees it, which is what "reveal …
// put that card into your hand" means.
//
// S22: "reveal" is now the shared primitive and covers the WHOLE run,
// not just the land. It used to mark every seat a knower of the basic
// land and nothing else, on the reasoning that the cards above it
// become public in the graveyard a moment later. They do, and that
// still left the table unable to tell a Hermit Druid activation from
// any other pile of cards arriving in a graveyard — which, for the
// card whose entire purpose is emptying a library in one activation,
// is the one thing worth announcing.
//
// With no basic land in the library every card is milled and the
// library is left empty — which is the combo the card is famous for.
// Nobody loses for the mill (CR 701.17b); the loss comes at the
// player's next draw (CR 704.5b), unless the combo wins first. That
// case reveals the whole library, as printed.
func b21RevealUntilBasicLandToHand(ctx *Context, player uuid.UUID) error {
	p := ctx.PlayerByID(player)
	if p == nil || p.Library == nil {
		return nil
	}
	above := 0
	land := uuid.Nil
	run := make([]uuid.UUID, 0, 8)
	for i := len(p.Library.Cards) - 1; i >= 0; i-- {
		run = append(run, p.Library.Cards[i].InstanceID)
		if IsBasicLand(p.Library.Cards[i]) {
			land = p.Library.Cards[i].InstanceID
			break
		}
		above++
	}
	if err := (RevealCards{
		Player: player,
		Cards:  run,
		Reason: "reveal from the top until a basic land card",
	}).Apply(ctx); err != nil {
		return err
	}
	if above > 0 {
		if err := (MillCards{Player: player, N: above}).Apply(ctx); err != nil {
			return err
		}
	}
	if land == uuid.Nil {
		return nil
	}
	return BounceToHand{Target: land}.Apply(ctx)
}

// b21ExileTopFourThenTakeTheirLands is Oblivion Sower's cast-trigger
// body: the targeted opponent exiles the top four cards of their
// library (an exile, not a mill — no mill payoff sees it), then every
// land card that player OWNS in exile — the four just exiled and any
// exiled earlier by anything else, as printed — is put onto the
// battlefield under the controller's control. The exile zone is
// walked once before the first move, because a return mutates it.
func b21ExileTopFourThenTakeTheirLands(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
		return nil
	}
	victim := item.Targets[0].ID
	return MillToZone{
		Player: victim,
		N:      4,
		To:     game.ZoneExile,
		// #893: the exile zone is walked from the continuation, because
		// the four cards are not all in it yet on the line after the
		// exile — a commander among them stops to answer CR 903.9, and
		// a card that takes the offer never reaches exile at all.
		Then: func(ctx *Context, _ []uuid.UUID) error {
			if ctx.Game.Exile == nil {
				return nil
			}
			var lands []uuid.UUID
			for _, c := range ctx.Game.Exile.Cards {
				if c.Owner == victim && c.IsLand() {
					lands = append(lands, c.InstanceID)
				}
			}
			for _, id := range lands {
				if err := (ReturnFromExile{Target: id, Controller: item.Controller}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	}.Apply(ctx)
}

// b21ReturnAllArtifactAndEnchantmentCards is Brilliant Restoration's
// body: every artifact and enchantment card in `player`'s graveyard
// returns to the battlefield under its owner's control. The IDs are
// snapshotted before the first move, because ReturnFromGraveyard
// mutates the pile being walked; each card enters through the
// ordinary reanimation path, so its own enters-tapped clause and
// every ETB trigger fire.
func b21ReturnAllArtifactAndEnchantmentCards(ctx *Context, player uuid.UUID) error {
	p := ctx.PlayerByID(player)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if c.IsArtifact() || c.IsEnchantment() {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
