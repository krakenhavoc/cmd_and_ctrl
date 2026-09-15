package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch34_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 34 (#397, `edhrec_rank` 3550–3649). Own file per
// the #231 convention; every package-level name carries the b34
// prefix because other batches land beside this one.
//
// What is NOT here, because main already had it: "this permanent
// enters" is b06SelfETB, "a land you control entered" is
// b33LandYouControlEntered, "this creature attacks" is
// attackDeclared and "a creature you control attacks" is
// attackDeclaredByYou, "a card left your graveyard" is
// b16CardLeftYourGraveyard with the per-label "one or more" dedup
// OncePerBatch, "you sacrificed a Food" is
// b31YouSacrificedAFood, "this died" is cardDied and "a creature
// died" is diedCreature, the historic test is b09IsHistoric, the
// lands-you-control count is b10LandsControlled, "each player draws
// N" is b05EachPlayerDraws, the draw-then-discard-choice loot is
// lootOne, "untap each X you control" is
// b16UntapAllYouControlMatching, "N damage to each opponent" is
// damageToEachOpponent, "target player draws and loses life" is
// b32TargetPlayerDrawsAndLosesLife, the first-legal-target read is
// b16FirstLegalTargetCard, "return the spell's graveyard targets to
// the battlefield" is b24ReturnGraveyardTargetsToBattlefield,
// "exile the top card until the end of your next turn" is
// b20ExileTopUntilEndOfNextTurn, the tapped 2/2 Zombie body is
// b33CreateTappedZombie, the Saproling is b11GreenSaprolingToken,
// the Goblin is RedGoblinToken, the Treasure and Food are in
// tokens.go, the basic-land test is b30IsBasicLandCard, "enters
// tapped unless you control a <type>" is
// SelfEntersTappedUnless(youControlLandTyped), the Background gate
// is b21ControlsCommanderCreatureYouOwn, and the lord builders are
// TribalAnthem / TribalKeywordGrant.

// --- tokens --------------------------------------------------------

// b34ScionOfTheDeepToken is Kiora, the Rising Tide's "Scion of the
// Deep, a legendary 8/8 blue Octopus creature token". Legendary is a
// real supertype on the token, so the legend rule applies to a
// second one exactly as printed.
func b34ScionOfTheDeepToken() game.Card {
	return game.Card{
		Name:      "Scion of the Deep",
		TypeLine:  "Token Legendary Creature — Octopus",
		Power:     8,
		Toughness: 8,
		Colors:    []string{"U"},
	}
}

// b34WhiteZombieToken is On Wings of Gold's 1/1 white Zombie. A
// Zombie AND a token, so the enchantment's own anthem lifts it.
func b34WhiteZombieToken() game.Card {
	return game.Card{
		Name:      "Zombie",
		TypeLine:  "Token Creature — Zombie",
		Power:     1,
		Toughness: 1,
		Colors:    []string{"W"},
	}
}

// --- card reads ----------------------------------------------------

// b34ElementalsControlled counts the Elementals `controller`
// controls — Omnath, Locus of the Roil's damage. Effective subtypes,
// so a changeling counts; Omnath himself is one and counts, as
// printed.
func b34ElementalsControlled(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.HasSubtype("Elemental") {
			n++
		}
	}
	return n
}

// b34CreaturesYouControlMatching counts the creatures `controller`
// controls that pass `match` — Armorcraft Judge's "creature you
// control with a +1/+1 counter on it", Regal Force's "green
// creature you control". Post-layer types and colours, so an
// animated artifact and a creature something painted green both
// count.
func b34CreaturesYouControlMatching(g *game.Game, controller uuid.UUID, match func(game.Card) bool) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && match(c) {
			n++
		}
	}
	return n
}

// b34HasPlusCounter is "with a +1/+1 counter on it".
func b34HasPlusCounter(c game.Card) bool { return c.Counters[game.CounterPlusOne] > 0 }

// b34IsGreen is "green creature" read off the effective colours.
func b34IsGreen(c game.Card) bool {
	for _, col := range c.EffectiveColors() {
		if col == "G" {
			return true
		}
	}
	return false
}

// b34OpponentsControlLandsAtLeast is Turbulent Fen's untapped
// condition: the controller's opponents, between them, control `n`
// or more lands. The reader takes the entering land's controller, as
// SelfEntersTappedUnless hands it.
func b34OpponentsControlLandsAtLeast(n int) func(g *game.Game, controller uuid.UUID) bool {
	return func(g *game.Game, controller uuid.UUID) bool {
		total := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller != controller && c.IsLand() {
				total++
			}
		}
		return total >= n
	}
}

// b34GraveyardHasAtLeast is threshold's test: `player`'s graveyard
// holds `n` or more cards.
func b34GraveyardHasAtLeast(g *game.Game, player uuid.UUID, n int) bool {
	p := g.PlayerByIDForEffect(player)
	return p != nil && p.Graveyard != nil && p.Graveyard.Size() >= n
}

// b34NoOpponentHasMoreLifeThan is Guild Artisan's intervening-if: no
// opponent of `controller` has a life total greater than `player`'s.
// The attacked player is normally one of those opponents and never
// has more life than themselves, so the question is about the
// others. A missing player answers false, which drops the trigger.
func b34NoOpponentHasMoreLifeThan(g *game.Game, controller, player uuid.UUID) bool {
	target := g.PlayerByIDForEffect(player)
	if target == nil {
		return false
	}
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == controller {
			continue
		}
		if p.Life > target.Life {
			return false
		}
	}
	return true
}

// b34YouSacrificedAFoodThisTurn is Elanor Gardner's end-step
// condition, walked off the event log back to the turn's upkeep
// (b06EnteredThisTurn's boundary): a sacrifice by `controller` of a
// permanent that is a Food. The sacrificed card is read wherever it
// now sits — a Food token in a graveyard keeps its type line — and
// one that can no longer be found does not count, which errs weaker.
func b34YouSacrificedAFoodThisTurn(g *game.Game, controller uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventSacrifice || ev.Actor != controller || ev.CardID == uuid.Nil {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok && c.HasSubtype("Food") {
			return true
		}
	}
	return false
}

// b34GoblinsEnteredUnderYourControlThisTurn is Hobgoblin Bandit
// Lord's count: the Goblins that entered the battlefield under
// `controller`'s control this turn, walked off the event log back to
// the turn's upkeep. Each entry is read wherever the card now sits,
// so a Goblin that has since died still counts, as printed; one
// that can no longer be found does not, which errs weaker. Effective
// subtypes, so a changeling counts.
func b34GoblinsEnteredUnderYourControlThisTurn(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventETB || ev.CardID == uuid.Nil {
			continue
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok || c.Controller != controller || !c.HasSubtype("Goblin") {
			continue
		}
		n++
	}
	return n
}

// b34DamageSourceController resolves who controls the source of a
// damage event. Combat damage carries the dealing creature's
// controller in Actor; spell and ability damage carries no Actor
// (only Source), so the source card is looked up wherever it is — a
// permanent on the battlefield, or a spell still on the stack while
// its effect deals the damage. A source that cannot be found is
// nobody's, which errs weaker.
func b34DamageSourceController(ev game.Event, g *game.Game) (uuid.UUID, bool) {
	if ev.Kind != game.EventDealDamage {
		return uuid.Nil, false
	}
	if ev.Combat && ev.Actor != uuid.Nil {
		return ev.Actor, true
	}
	if ev.Source == uuid.Nil {
		return uuid.Nil, false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	if !ok {
		return uuid.Nil, false
	}
	return c.Controller, true
}

// --- trigger conditions --------------------------------------------

// b34LandAnOpponentControlsEntered is Polluted Bonds' condition: a
// land entered the battlefield under the control of an opponent of
// the source's controller. Returns the land. Post-layer types, so
// an animated land still counts and a Dryad Arbor is a land.
func b34LandAnOpponentControlsEntered(ev game.Event, source *game.Card, g *game.Game) (game.Card, bool) {
	if ev.Kind != game.EventETB || ev.CardID == uuid.Nil {
		return game.Card{}, false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsLand() || c.Controller == source.Controller {
		return game.Card{}, false
	}
	if p := g.PlayerByIDForEffect(c.Controller); p == nil || p.Eliminated {
		return game.Card{}, false
	}
	return c, true
}

// b34SelfOrAnotherNontokenHistoricYouControlEntered is Arbaaz Mir's
// condition: Arbaaz himself entered, or another nontoken historic
// permanent — an artifact, a legendary, a Saga — entered under his
// controller's control. Historic is read post-layer (b09IsHistoric),
// so an artifact something else animated still counts and a token
// copy of a legend does not.
func b34SelfOrAnotherNontokenHistoricYouControlEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind == game.EventETB && ev.CardID == source.InstanceID {
		return true
	}
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && !IsToken(c) && b09IsHistoric(c)
}

// b34CommanderCreatureYouOwnAttackedAPlayer is Guild Artisan's
// condition before its intervening-if: a creature that is a
// commander the source's controller OWNS, and currently controls,
// was declared as an attacker against a player. "You own" is the
// printed clause; control is what makes the granted ability exist
// on the Background's side (see the card file).
func b34CommanderCreatureYouOwnAttackedAPlayer(ev game.Event, source *game.Card, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsCommander || c.Owner != source.Controller || !c.IsCreature() {
		return false
	}
	return g.ClassifyAttackTargetForEffect(ev.Target) == game.AttackTargetPlayer
}

// b34VampireYouControlAttacked is Crossway Troublemakers' first
// condition: a Vampire the source's controller controls was declared
// as an attacker — the Troublemakers themselves included. Effective
// subtypes, so a changeling counts.
func b34VampireYouControlAttacked(ev game.Event, source *game.Card, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.HasSubtype("Vampire")
}

// b34VampireYouControlDied is Crossway Troublemakers' second
// condition: a Vampire creature its controller controlled died —
// the Troublemakers' own death included, since a dies trigger looks
// back (CR 603.10) and the card in the graveyard still reads as a
// Vampire.
func b34VampireYouControlDied(ev game.Event, source *game.Card, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	return ok && dead.Controller == source.Controller && dead.HasSubtype("Vampire")
}

// b34SourceYouControlDealtDamageToPlayerAtLeast is Dragonborn
// Champion's condition: one damage event of `n` or more from a
// source the source's controller controls to a player. One event
// per source per recipient, as the engine emits them, so two
// creatures connecting for 3 each do not add up to 5 — as printed.
func b34SourceYouControlDealtDamageToPlayerAtLeast(ev game.Event, source *game.Card, g *game.Game, n int) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount < n || g.PlayerByIDForEffect(ev.Target) == nil {
		return false
	}
	controller, ok := b34DamageSourceController(ev, g)
	return ok && controller == source.Controller
}

// --- statics -------------------------------------------------------

// b34ZombieOrTokenYouControl is On Wings of Gold's scope: a creature
// the source's controller controls that is a Zombie and/or a token.
func b34ZombieOrTokenYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.Controller == source.Controller &&
		(target.HasSubtype("Zombie") || IsToken(*target))
}

// b34ZombiesAndTokensYouControlGetPlusOne is the +1/+1 half of On
// Wings of Gold — layer 7c.
func b34ZombiesAndTokensYouControlGetPlusOne() game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: b34ZombieOrTokenYouControl,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power++
			c.Toughness++
		},
	}
}

// b34ZombiesAndTokensYouControlHaveFlying is the flying half of On
// Wings of Gold — layer 6, deduped.
func b34ZombiesAndTokensYouControlHaveFlying() game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		AppliesTo: b34ZombieOrTokenYouControl,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			for _, k := range c.Abilities {
				if k == "flying" {
					return
				}
			}
			c.Abilities = append(c.Abilities, "flying")
		},
	}
}

// --- replacements --------------------------------------------------

// b34DoubleDamageToOpponents is Fiendish Duo's "If a source would
// deal damage to an opponent, it deals double that damage to that
// player instead": every source, combat and noncombat, the
// controller's own included, but only when the recipient is a
// PLAYER who is an opponent of the Duo's controller. Damage to a
// permanent, or to the controller, is untouched.
func b34DoubleDamageToOpponents(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 {
				return false
			}
			p := g.PlayerByIDForEffect(ev.DamageTarget)
			return p != nil && !p.Eliminated && p.ID != src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.DamageAmount *= 2
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// --- triggers ------------------------------------------------------

// b34AttackingVampiresHaveDeathtouchAndLifelink is Crossway
// Troublemakers' printed static written as an attack trigger, for
// the reason Blade Historian gives: an attack declaration does not
// rebuild the layer cache, so a static gated on "is attacking"
// would not see it. One trigger per attacking Vampire the
// controller controls, granting both keywords until end of turn
// through the turn-scoped registry, which does invalidate the cache.
func b34AttackingVampiresHaveDeathtouchAndLifelink(name string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventAttack},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return b34VampireYouControlAttacked(ev, source, g)
		},
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			attacker := ev.CardID
			return game.NewTriggeredItem(source, name+" — the attacking Vampire has deathtouch and lifelink",
				func(g *game.Game, item *game.StackItem) error {
					return GrantKeywordUntilEOT{
						Target:   attacker,
						Keywords: []string{"deathtouch", "lifelink"},
						Label:    name + " — deathtouch and lifelink",
					}.Apply(NewContext(g, item))
				})
		},
	}
}

// --- effect bodies -------------------------------------------------

// b34PutCounterThenDoubleCounters is Invigorating Surge's body: one
// +1/+1 counter on the announced creature, if still legal, then as
// many more as it now has — "double the number of +1/+1 counters on
// that creature", in printed order, so a creature with none ends on
// two. Two placements, so a Hardened Scales sees both.
func b34PutCounterThenDoubleCounters(ctx *Context) error {
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	c, found := ctx.Game.LookupCardForEffect(id)
	if !found || !onBattlefield(ctx.Game, id) {
		return nil
	}
	n := c.Counters[game.CounterPlusOne]
	if n <= 0 {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
}

// b34DrawPerCreatureYouControlMatching is "draw a card for each
// creature you control that <match>" as a trigger body — Armorcraft
// Judge, Regal Force. Counted at resolution, as printed.
func b34DrawPerCreatureYouControlMatching(match func(game.Card) bool) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		n := b34CreaturesYouControlMatching(g, item.Controller, match)
		return DrawCards{Player: item.Controller, N: n}.Apply(NewContext(g, item))
	}
}

// b34DamageChosenTargetPerElemental is Omnath, Locus of the Roil's
// entry body: damage from Omnath to the announced target, if still
// legal, equal to the Elementals his controller controls as the
// trigger resolves. Omnath removed in response deals nothing for
// himself but still counts every other Elemental — the source of
// the damage is Omnath by last known information.
func b34DamageChosenTargetPerElemental(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := b34ElementalsControlled(g, item.Controller)
	for _, t := range ctx.LegalTargets() {
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
	}
	return nil
}

// b34CounterOnChosenElementalThenDrawAtEightLands is Omnath's
// landfall body: a +1/+1 counter on the announced Elemental, if
// still legal, then a card if the controller controls eight or more
// lands as the trigger resolves. A target that left in response
// counters the whole ability (CR 608.2b), the draw included.
func b34CounterOnChosenElementalThenDrawAtEightLands(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if b10LandsControlled(g, item.Controller) < 8 {
		return nil
	}
	return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
}

// b34SearchUpToTwoBasicsToHand is Yavimaya Elder's dies body: up to
// two basic land cards, revealed, to hand, then shuffle. The "may"
// was answered when the trigger fired; the search prompt is where
// the controller picks which two, or fewer.
func b34SearchUpToTwoBasicsToHand(g *game.Game, item *game.StackItem) error {
	return SearchLibrary{
		Player:    item.Controller,
		Predicate: b30IsBasicLandCard,
		Dest:      game.ZoneHand,
		Limit:     2,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "Yavimaya Elder — up to two basic land cards to put into your hand",
	}.Apply(NewContext(g, item))
}

// b34ReturnUpToTwoSmallCreatures is Reveillark's leave body: the
// announced creature cards, up to two, still in the graveyard and
// still power 2 or less, return to the battlefield under their
// owner's control — the controller's own graveyard, so the owner is
// the controller.
func b34ReturnUpToTwoSmallCreatures(g *game.Game, item *game.StackItem) error {
	return b24ReturnGraveyardTargetsToBattlefield(NewContext(g, item), 2)
}

// b34DestroyChosenAndAllOthersWithItsName is Maelstrom Pulse's body:
// the announced nonland permanent, if still legal, and every other
// permanent on the battlefield that shares its name — everyone's,
// lands included ("all other permanents"), tokens included. Each is
// destroyed through the single-target verb, so an indestructible one
// survives while the rest go. The name is read at resolution, off
// the effective characteristics.
func b34DestroyChosenAndAllOthersWithItsName(ctx *Context) error {
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	chosen, found := ctx.Game.LookupCardForEffect(id)
	if !found {
		return nil
	}
	name := chosen.Effective().Name
	ids := []uuid.UUID{id}
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.InstanceID != id && c.Effective().Name == name {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, target := range ids {
		if !onBattlefield(ctx.Game, target) {
			continue
		}
		if err := (DestroyTarget{Target: target}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b34ThatPlayerLosesLifeAndDraws is Seizan, Perverter of Truth's
// upkeep body: the player whose upkeep it is — captured in Build —
// loses `life` life and draws `draw` cards, in printed order. A
// player no longer seated does neither.
func b34ThatPlayerLosesLifeAndDraws(player uuid.UUID, life, draw int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if p := g.PlayerByIDForEffect(player); p == nil || p.Eliminated {
			return nil
		}
		if err := g.ChangePlayerLifeForEffect(item.SourceCardID, player, -life); err != nil {
			return err
		}
		return DrawCards{Player: player, N: draw}.Apply(NewContext(g, item))
	}
}

// b34PayLifeToDraw is Crossway Troublemakers' dies body: the
// controller pays `life` life and draws a card. The "you may" was
// answered when the trigger fired; a controller who can no longer
// pay (CR 119.4 — a life payment needs at least that much life)
// neither pays nor draws.
func b34PayLifeToDraw(life int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		p := g.PlayerByIDForEffect(item.Controller)
		if p == nil || p.Eliminated || p.Life < life {
			return nil
		}
		if err := g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -life); err != nil {
			return err
		}
		return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
	}
}

// b34DamageEachOpponentAndGainLife is Arbaaz Mir's body: `damage`
// from Arbaaz to each opponent, then the controller gains `life`.
func b34DamageEachOpponentAndGainLife(damage, life int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if err := damageToEachOpponent(g, item, damage); err != nil {
			return err
		}
		return GainLife{Player: item.Controller, Amount: life}.Apply(NewContext(g, item))
	}
}

// b34DamageChosenTargetPerGoblinEnteredThisTurn is Hobgoblin Bandit
// Lord's activation body: damage from the Lord to the announced
// target, if still legal, equal to the Goblins that entered under
// the controller's control this turn, counted as the ability
// resolves.
func b34DamageChosenTargetPerGoblinEnteredThisTurn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := b34GoblinsEnteredUnderYourControlThisTurn(g, item.Controller)
	for _, t := range ctx.LegalTargets() {
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
	}
	return nil
}

// b34SearchBasicTappedIfSacrificedAFood is Elanor Gardner's end-step
// body: the intervening-if re-checked at resolution (CR 603.4), then
// a basic land card onto the battlefield tapped and a shuffle. The
// "may" was answered when the trigger fired.
func b34SearchBasicTappedIfSacrificedAFood(g *game.Game, item *game.StackItem) error {
	if !b34YouSacrificedAFoodThisTurn(g, item.Controller) {
		return nil
	}
	return SearchLibrary{
		Player:        item.Controller,
		Predicate:     b30IsBasicLandCard,
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		TappedOnEntry: true,
		Shuffle:       true,
		Reason:        "Elanor Gardner — a basic land card to put onto the battlefield tapped",
	}.Apply(NewContext(g, item))
}

// b34ThatPlayerLosesTwoYouGainTwo is Polluted Bonds' body: the
// land's controller — captured in Build — loses 2 life, then the
// enchantment's controller gains 2. A loss and a gain, not damage
// and not a drain of whatever was lost: the gain is 2 even when the
// loser is already gone.
func b34ThatPlayerLosesTwoYouGainTwo(victim uuid.UUID) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if p := g.PlayerByIDForEffect(victim); p != nil && !p.Eliminated {
			if err := g.ChangePlayerLifeForEffect(item.SourceCardID, victim, -2); err != nil {
				return err
			}
		}
		return GainLife{Player: item.Controller, Amount: 2}.Apply(NewContext(g, item))
	}
}

// b34DestroyChosenThenSearchBasicTapped is Deathsprout's body: the
// announced creature is destroyed, then a basic land card is put
// onto the battlefield tapped and the library shuffled — the search
// is not conditional on the destroy, as printed, so an
// indestructible creature still ramps the caster. A creature that
// left in response never reaches here: the engine fizzles a
// single-target spell before its body runs (CR 608.2b).
func b34DestroyChosenThenSearchBasicTapped(ctx *Context) error {
	if id, ok := b16FirstLegalTargetCard(ctx); ok {
		if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return SearchLibrary{
		Player:        ctx.Controller(),
		Predicate:     b30IsBasicLandCard,
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		TappedOnEntry: true,
		Shuffle:       true,
		Reason:        "Deathsprout — a basic land card to put onto the battlefield tapped",
	}.Apply(ctx)
}

// b34IsBasicLandOrDesertCard is Map the Frontier's predicate: a basic
// land card, or any card with the Desert land type.
func b34IsBasicLandOrDesertCard(c game.Card) bool {
	return b30IsBasicLandCard(c) || (c.IsLand() && c.HasSubtype("Desert"))
}

// b34SearchUpToTwoBasicsOrDesertsTapped is Map the Frontier's body:
// up to two cards, each a basic land or a Desert, onto the
// battlefield tapped, then shuffle.
func b34SearchUpToTwoBasicsOrDesertsTapped(ctx *Context) error {
	return SearchLibrary{
		Player:        ctx.Controller(),
		Predicate:     b34IsBasicLandOrDesertCard,
		Dest:          game.ZoneBattlefield,
		Limit:         2,
		TappedOnEntry: true,
		Shuffle:       true,
		Reason:        "Map the Frontier — up to two basic land cards and/or Desert cards to put onto the battlefield tapped",
	}.Apply(ctx)
}

// b34CreateScionIfThreshold is Kiora, the Rising Tide's attack body:
// threshold re-checked at resolution (CR 603.4), then Scion of the
// Deep. The "may" was answered when the trigger fired.
func b34CreateScionIfThreshold(g *game.Game, item *game.StackItem) error {
	if !b34GraveyardHasAtLeast(g, item.Controller, 7) {
		return nil
	}
	return CreateToken{Controller: item.Controller, Template: b34ScionOfTheDeepToken(), N: 1}.Apply(NewContext(g, item))
}

// b34UntapAllCreaturesYouControl is "untap all creatures you
// control" — Vitalize's resolution and Village Bell-Ringer's entry.
func b34UntapAllCreaturesYouControl(g *game.Game, item *game.StackItem) error {
	return b16UntapAllYouControlMatching(NewContext(g, item), item.Controller, func(c game.Card) bool { return c.IsCreature() })
}

// b34CreateTokens is "create N <token>s" as a trigger body.
func b34CreateTokens(template func() game.Card, n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		return CreateToken{Controller: item.Controller, Template: template(), N: n}.Apply(NewContext(g, item))
	}
}

// b34ReturnChosenGraveyardCardToHand is Auroral Procession's body
// and Lord of the Undead's activation: the announced card, if still
// in a graveyard, to its owner's hand.
func b34ReturnChosenGraveyardCardToHand(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
	}
	return nil
}
