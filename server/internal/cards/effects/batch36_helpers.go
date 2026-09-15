package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch36_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 36 (#399, `edhrec_rank` 3753–3852). Own file per
// the #231 convention; every package-level name carries the b36
// prefix because other batches land beside this one.
//
// What is NOT here, because main already had it: "this permanent
// enters" is b06SelfETB, "another creature you control entered" is
// b13AnotherCreatureYouControlEntered, "an Angel you control
// entered" is b21AngelYouControlEntered, "a creature you control
// attacks" is attackDeclaredByYou with the "one or more" dedup
// OncePerBatch, "a creature you control dealt combat
// damage to a player" is combatDamageToPlayerBy, "a creature died"
// is diedCreature, "an opponent lost N or more life this turn" is
// b06AnOpponentLostAtLeastThisTurn, "each end step" is
// b15EndStepBegan, devotion is devotionTo, "N damage to each
// opponent" is damageToEachOpponent, "the permanent sacrificed to
// pay" is b17PermanentSacrificedToPay, the tutor-to-hand body is
// b06TutorToHand, the first-legal-target read is
// b16FirstLegalTargetCard, "the cards exiled with this permanent"
// is b27ExiledWith, "another" by name is b03NotNamed, the basic-land
// test is b30IsBasicLandCard, "enters tapped unless your opponents
// control N lands" is SelfEntersTappedUnless(
// b34OpponentsControlLandsAtLeast(n)), the white flying Spirit is
// b28WhiteSpiritFlyingToken, the predicate-scoped anthem is
// b16Anthem, and the lord builders are TribalAnthem /
// TribalKeywordGrant.

// --- tokens --------------------------------------------------------

// b36WhiteCatToken is Arahbo, the First Fang's 1/1 white Cat. A Cat,
// so Arahbo's own anthem lifts it; a token, so it never fires his
// nontoken-Cat trigger again.
func b36WhiteCatToken() game.Card {
	return game.Card{
		Name:      "Cat",
		TypeLine:  "Token Creature — Cat",
		Power:     1,
		Toughness: 1,
		Colors:    []string{"W"},
	}
}

// b36WhiteVampireLifelinkToken is Mavren Fein, Dusk Apostle's 1/1
// white Vampire with lifelink. A token, so it never fires the
// nontoken-Vampires-attack trigger by itself.
func b36WhiteVampireLifelinkToken() game.Card {
	return game.Card{
		Name:      "Vampire",
		TypeLine:  "Token Creature — Vampire",
		Power:     1,
		Toughness: 1,
		Colors:    []string{"W"},
		Keywords:  []string{"lifelink"},
	}
}

// b36WhiteRabbitToken is Cadira, Caller of the Small's 1/1 white
// Rabbit. Vanilla; a token, so the next hit counts it.
func b36WhiteRabbitToken() game.Card {
	return game.Card{
		Name:      "Rabbit",
		TypeLine:  "Token Creature — Rabbit",
		Power:     1,
		Toughness: 1,
		Colors:    []string{"W"},
	}
}

// b36BlueThopterToken is Sharding Sphinx's 1/1 blue Thopter artifact
// creature with flying — tokens.go's ThopterToken is colourless, and
// the colour is printed. An artifact creature, so the Thopter itself
// grows the next batch.
func b36BlueThopterToken() game.Card {
	return game.Card{
		Name:      "Thopter",
		TypeLine:  "Token Artifact Creature — Thopter",
		Power:     1,
		Toughness: 1,
		Colors:    []string{"U"},
		Keywords:  []string{"flying"},
	}
}

// b36WhiteHumanSoldierToken is Lossarnach Captain's 1/1 white Human
// Soldier — tokens.go's HumanSoldierToken is colourless. A Human, so
// the Captain's own tap trigger fires when it enters.
func b36WhiteHumanSoldierToken() game.Card {
	return game.Card{
		Name:      "Human Soldier",
		TypeLine:  "Token Creature — Human Soldier",
		Power:     1,
		Toughness: 1,
		Colors:    []string{"W"},
	}
}

// --- costs ---------------------------------------------------------

// b36SacrificeAnotherArtifact is Repurposing Bay's "Sacrifice another
// artifact:" — any artifact the activator controls other than one
// named `self`. Built through the same clause the sac-outlet costs
// use, so the client opens the same picker; "another" by name, the
// b03NotNamed posture, since the clause is declared before any
// instance exists.
func b36SacrificeAnotherArtifact(self string) game.AbilityCost {
	return game.AbilityCost{SacrificeOther: sacrificeSpec("another artifact", Artifact(), b03NotNamed(self))}
}

// --- predicates ----------------------------------------------------

// b36CommanderYouOwn is Sanctum of Eternity's "target commander you
// own": a permanent flagged as a commander whose OWNER is the
// activator — a stolen commander of yours is still yours to return,
// and an opponent's commander you control is not.
func b36CommanderYouOwn() CardPredicate {
	return func(_ *game.Game, caster uuid.UUID, c game.Card) bool {
		return c.IsCommander && c.Owner == caster
	}
}

// b36ArtifactCreatureYouControl is Tempered Steel's scope — post-
// layer types, so an animated artifact and an artifact-ified
// creature both count.
func b36ArtifactCreatureYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.IsArtifact() && target.Controller == source.Controller
}

// b36IsAuraCard is Heliod's Pilgrim's search filter: an Aura card,
// read off the printed type line as every library search does.
func b36IsAuraCard(c game.Card) bool { return c.IsAura() }

// b36IsDragonCreatureCard is Sarkhan's Triumph's search filter: a
// creature card with the Dragon type, on the printed type line.
func b36IsDragonCreatureCard(c game.Card) bool {
	return c.IsCreature() && c.HasSubtype("Dragon")
}

// --- card reads ----------------------------------------------------

// b36TokensControlled counts the tokens `controller` controls, of
// any type — Cadira's "for each token you control". A Treasure and
// a Clue count, as printed.
func b36TokensControlled(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && IsToken(c) {
			n++
		}
	}
	return n
}

// --- trigger conditions --------------------------------------------

// b36SelfOrAnotherNontokenCatYouControlEntered is Arahbo, the First
// Fang's condition: Arahbo himself entered, or another nontoken Cat
// entered under his controller's control. Effective subtypes, so a
// changeling counts; his own Cat tokens never do.
func b36SelfOrAnotherNontokenCatYouControlEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind == game.EventETB && ev.CardID == source.InstanceID {
		return true
	}
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.IsCreature() && !IsToken(c) && c.HasSubtype("Cat")
}

// b36SelfOrAnotherHumanYouControlEntered is Lossarnach Captain's
// condition: the Captain entered, or another Human — token or not —
// entered under his controller's control. His own Human Soldier
// tokens count, as printed.
func b36SelfOrAnotherHumanYouControlEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind == game.EventETB && ev.CardID == source.InstanceID {
		return true
	}
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.HasSubtype("Human")
}

// b36NontokenVampiresYouControlAttacked is Mavren Fein, Dusk
// Apostle's "whenever one or more nontoken Vampires you control
// attack": a nontoken Vampire the source's controller controls was
// declared as an attacker — Mavren himself included — deduplicated
// per combat through OncePerBatch, because the engine
// emits one attack event per creature and the printed ability fires
// once for the batch.
func b36NontokenVampiresYouControlAttacked(ev game.Event, source *game.Card, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && !IsToken(c) && c.HasSubtype("Vampire")
}

// b36ArtifactCreatureYouControlDealtCombatDamageToPlayer is Sharding
// Sphinx's condition: a creature the controller controls that is
// also an artifact — the Sphinx itself, a Thopter, an animated
// Treasure — dealt combat damage to a player. One trigger per
// creature that connects, as printed.
func b36ArtifactCreatureYouControlDealtCombatDamageToPlayer(ev game.Event, source *game.Card, g *game.Game) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && c.IsArtifact()
}

// b36SelfDealtCombatDamageToPlayer is "whenever this creature deals
// combat damage to a player" — Cadira.
func b36SelfDealtCombatDamageToPlayer(ev game.Event, source *game.Card, g *game.Game) bool {
	return ev.Source == source.InstanceID && combatDamageToPlayerBy(ev, source.Controller, g)
}

// b36AngelYouControlDied is Bishop of Wings' second condition: an
// Angel its controller controlled died. The card in the graveyard
// still reads as an Angel; a changeling counts.
func b36AngelYouControlDied(ev game.Event, source *game.Card, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	return ok && dead.Controller == source.Controller && dead.HasSubtype("Angel")
}

// b36AnotherFaerieYouControlDied is Tegwyll, Duke of Splendor's
// condition: another Faerie his controller controlled died —
// Tegwyll's own death does not count, as printed.
func b36AnotherFaerieYouControlDied(ev game.Event, source *game.Card, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	return ok && dead.InstanceID != source.InstanceID &&
		dead.Controller == source.Controller && dead.HasSubtype("Faerie")
}

// b36EndStepAndAnOpponentLostThree is Sygg, River Cutthroat's
// intervening-if at announce (CR 603.4): any player's end step
// began, and some one opponent of the controller has lost 3 or more
// life this turn.
func b36EndStepAndAnOpponentLostThree(ev game.Event, source *game.Card, g *game.Game) bool {
	return b15EndStepBegan(ev) && b06AnOpponentLostAtLeastThisTurn(g, source.Controller, 3)
}

// --- effect bodies -------------------------------------------------

// b36AngelOfSerenityExileLabel is the stack label of Angel of
// Serenity's entry trigger — the "exiled with" record keys on it.
const b36AngelOfSerenityExileLabel = "Angel of Serenity — exile up to three other creatures"

// b36ExileChosenCreatures is Angel of Serenity's entry body: every
// announced creature still legal at resolution is exiled, the Angel
// herself excepted. The "may" was answered when the trigger fired.
func b36ExileChosenCreatures(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || t.ID == item.SourceCardID {
			continue
		}
		if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b36ReturnCardsExiledWithToOwnersHands is Angel of Serenity's leave
// body: every card still in exile that her entry trigger exiled
// returns to its owner's hand. A card that has since left exile is
// not counted (b27ExiledWith closes the record), so nothing is
// pulled out of a graveyard or a hand.
func b36ReturnCardsExiledWithToOwnersHands(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range b27ExiledWith(g, item.SourceCardID, b36AngelOfSerenityExileLabel) {
		if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b36ReturnAllNonAuraEnchantmentCards is Replenish's body: every
// enchantment card in the caster's graveyard that is not an Aura
// returns to the battlefield under its owner's control, each through
// the ordinary reanimation path so its ETB triggers fire. Auras stay
// where they are — the printed "Auras with nothing to enchant remain
// in your graveyard", applied to every Aura, because the reanimation
// path has no CR 303.4f choose-what-to-enchant prompt (see the card
// file). The IDs are snapshotted before the first move, because the
// pile being walked mutates.
func b36ReturnAllNonAuraEnchantmentCards(ctx *Context) error {
	p := ctx.PlayerByID(ctx.Controller())
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if c.IsEnchantment() && !c.IsAura() {
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

// b36RepurposingBaySearch is Repurposing Bay's body — Oswald
// Fiddlebender's, one mana over: the sacrificed artifact's mana
// value is read back off the event log, and the search accepts only
// artifact cards at exactly that value plus one, put onto the
// battlefield untapped.
func b36RepurposingBaySearch(g *game.Game, item *game.StackItem) error {
	sacrificed, ok := b17PermanentSacrificedToPay(g, item)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(sacrificed)
	if !ok {
		return nil
	}
	want := c.ManaValue() + 1
	return SearchLibrary{
		Player: item.Controller,
		Predicate: func(c game.Card) bool {
			return c.IsArtifact() && c.ManaValue() == want
		},
		Dest:    game.ZoneBattlefield,
		Limit:   1,
		Shuffle: true,
		Reason:  "Repurposing Bay — an artifact card with mana value " + strconv.Itoa(want),
	}.Apply(NewContext(g, item))
}

// b36BounceChosenCommander is Sanctum of Eternity's body: the
// announced commander, if still on the battlefield and still legal,
// goes to its owner's hand — through the shared exit primitive, so
// CR 903.9 offers the command zone on the way (#539).
func b36BounceChosenCommander(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok || !onBattlefield(g, id) {
		return nil
	}
	return BounceToHand{Target: id}.Apply(ctx)
}

// b36CounterOnChosenAnimal is Animal Sanctuary's body: a +1/+1
// counter on the announced permanent, if still legal.
func b36CounterOnChosenAnimal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}

// b36TapChosenCreature is Lossarnach Captain's entry body: the
// announced creature, if still legal, is tapped.
func b36TapChosenCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return TapTarget{Target: id}.Apply(ctx)
}

// b36DamageEachOpponentForDevotionToRed is Fanatic of Mogis' entry
// body: damage from the Fanatic to each opponent equal to the
// controller's devotion to red, counted as the trigger resolves —
// the Fanatic's own {R} included while he is still there.
func b36DamageEachOpponentForDevotionToRed(g *game.Game, item *game.StackItem) error {
	n := devotionTo(g, item.Controller, "R")
	if n <= 0 {
		return nil
	}
	return damageToEachOpponent(g, item, n)
}

// b36RabbitsPerTokenYouControl is Cadira's body: one 1/1 white
// Rabbit for each token the controller controls as the trigger
// resolves — the Rabbits enter together, so none of them counts
// itself.
func b36RabbitsPerTokenYouControl(g *game.Game, item *game.StackItem) error {
	n := b36TokensControlled(g, item.Controller)
	if n <= 0 {
		return nil
	}
	return CreateToken{Controller: item.Controller, Template: b36WhiteRabbitToken(), N: n}.Apply(NewContext(g, item))
}

// b36DrawAndLoseOne is Tegwyll's dies body: the controller draws a
// card, then loses 1 life, in printed order.
func b36DrawAndLoseOne(g *game.Game, item *game.StackItem) error {
	if err := (DrawCards{Player: item.Controller, N: 1}).Apply(NewContext(g, item)); err != nil {
		return err
	}
	return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1)
}

// b36DrawIfAnOpponentLostThree is Sygg's end-step body: the
// intervening-if re-checked at resolution (CR 603.4), then a card.
// The "may" was answered when the trigger fired.
func b36DrawIfAnOpponentLostThree(g *game.Game, item *game.StackItem) error {
	if !b06AnOpponentLostAtLeastThisTurn(g, item.Controller, 3) {
		return nil
	}
	return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
}

// b36GainLife is "you gain N life" as a trigger body — Lifecreed
// Duo's 1, Bishop of Wings' 4.
func b36GainLife(n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		return GainLife{Player: item.Controller, Amount: n}.Apply(NewContext(g, item))
	}
}

// b36DrawOne is "draw a card" as a trigger body — Llanowar
// Visionary's entry.
func b36DrawOne(g *game.Game, item *game.StackItem) error {
	return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
}

// b36BounceChosenCreature is Whitemane Lion's entry body: the
// announced creature — the Lion itself included — returns to its
// owner's hand if it is still on the battlefield.
func b36BounceChosenCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok || !onBattlefield(g, id) {
		return nil
	}
	return BounceToHand{Target: id}.Apply(ctx)
}

// b36DesertDual is the Outlaws of Thunder Junction tapped Desert
// dual — Abraded Bluffs' and Bristling Backwoods' shape:
//
//	"This land enters tapped.
//	 When this land enters, it deals 1 damage to target opponent.
//	 {T}: Add {A} or {B}."
//
// The tapped entry is the real CR 614 self-replacement; the enters
// trigger is targeted ("target opponent") and dealt by the land, so
// it is colourless noncombat damage; the two printed colours are not
// narrowed to the commander's identity (the painland posture).
func b36DesertDual(oracleID, name, a, b string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   TargetPlayer("target opponent", Opponent()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, name+" — 1 damage to target opponent",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
							return nil
						}
						return DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: 1}.Apply(NewContext(g, item))
					})
			},
		}},
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true},
			Produced:                "{" + a + "|" + b + "}",
			Label:                   "Add {" + a + "} or {" + b + "}",
			IgnoreCommanderIdentity: true,
		}},
	}
}
