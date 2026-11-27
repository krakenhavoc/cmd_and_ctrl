package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b579_test.go — Ravnica: City of Guilds, batch 579 (#5600). Every test
// plays the real catalog entry through the engine.

const (
	oldB579Eye        = "01f8db27-1191-488e-b0d5-172821613c15"
	oldB579Glaze      = "36a247d6-0d40-4f61-935c-60f42aa98562"
	oldB579Field      = "63439f19-01e5-4ae1-9841-90493d62d0c8"
	oldB579Funnel     = "81dd22f7-f6c6-4cf2-9744-b1d918dbcf9a"
	oldB579Light      = "c34cf404-729d-46ef-8661-1095e5581766"
	oldB579Smash      = "1602f8c7-fe00-4bc9-b4fc-b26b7223ac75"
	oldB579Mindmoil   = "41e4ba67-d038-4d0a-b740-709cb2350192"
	oldB579Quick      = "77f34281-6f93-4043-a13d-8d83014e7620"
	oldB579Justice    = "89c5dec1-cefe-4ce3-a9df-b5cb9fa73b35"
	oldB579Festival   = "f8d1fbec-6ec1-47fa-ba02-fcee906db0d7"
	oldB579Bolt       = "4457ed35-7c10-48c8-9776-456485fdf070"
	oldB579Divination = "273b339c-964b-4a18-8eb5-ceb8abcdfd9e"
	oldB579Lizard     = "667dae0b-5006-4e38-be80-eae79c2a2a95"
	oldB579Elves      = "68954295-54e3-4303-a6bc-fc4547a4e3a3"
	oldB579Bear       = "Creature — Bear"
)

func TestOldB579EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB579Eye: "Eye of the Storm", oldB579Glaze: "Halcyon Glaze", oldB579Field: "Suppression Field",
		oldB579Funnel: "Blood Funnel", oldB579Light: "Light of Sanction", oldB579Smash: "Smash",
		oldB579Mindmoil: "Mindmoil", oldB579Quick: "Quickchange", oldB579Justice: "Twisted Justice",
		oldB579Festival: "Festival of the Guildpact",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB579SmashDestroysAnArtifactAndDrawsACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.StockLibraryFor(g, me, 5)
	rock := effectstest.B12Push(g, opp.ID, "Mind Stone", "Artifact", "", 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB579Bear, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Smash", "Instant", oldB579Smash, effectstest.CardTarget(bear)); err == nil {
		t.Fatal("Smash accepted a creature that is not an artifact")
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Smash", "Instant", oldB579Smash, effectstest.CardTarget(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d (one card drawn)", got, hand+1)
	}
}

func TestOldB579HalcyonGlazeBecomesAFlyingIllusionWhenYouCastACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	glaze := effectstest.B12Push(g, me.ID, "Halcyon Glaze", "Enchantment", oldB579Glaze, 0, 0)
	spell := effectstest.HandCardFull(me, "Shock", "Instant", "{R}", "", []string{"R"})
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.PlayerRefs(g.Seats[1].ID)}); err != nil {
		t.Logf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LayeredCard(t, g, glaze).IsCreature() {
		t.Fatal("a noncreature spell animated the Glaze")
	}
	bear := effectstest.HandCardFull(me, "Grizzly Bears", oldB579Bear, "{1}{G}", "", []string{"G"})
	effectstest.FloatMana(t, g, me, "{G}{G}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, glaze)
	if !c.IsCreature() || !c.HasSubtype("Illusion") || !c.HasCardType("enchantment") {
		t.Errorf("the Glaze is %q, want an Illusion enchantment creature", c.TypeLine)
	}
	if p, tt := effectstest.PtOf(t, g, glaze); p != 4 || tt != 4 {
		t.Errorf("the Glaze is %d/%d, want 4/4", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, glaze, "flying") {
		t.Error("the Glaze has no flying")
	}
	effectstest.RunCleanup(t, g)
	if effectstest.LayeredCard(t, g, glaze).IsCreature() {
		t.Error("the Glaze is still a creature after the turn")
	}
}

func TestOldB579SuppressionFieldTaxesActivatedAbilitiesButNotManaAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, g.Seats[1].ID, "Suppression Field", "Enchantment", oldB579Field, 0, 0)
	lizard := effectstest.B12Push(g, me.ID, "Bellows Lizard", "Creature — Lizard", oldB579Lizard, 1, 1)
	effectstest.FloatMana(t, g, me, "{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, lizard, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated the {1}{R} ability for {1}{R} under Suppression Field")
	}
	effectstest.FloatMana(t, g, me, "{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, lizard, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate for {3}{R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, lizard); p != 2 {
		t.Errorf("power %d, want 2", p)
	}
	elves := effectstest.B12Push(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", oldB579Elves, 1, 1)
	me.ManaPool = game.ManaPool{}
	if err := g.ActivateManaAbility(me.ID, elves, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("the mana ability was taxed or refused: %v", err)
	}
}

func oldB579PayPrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		if c := g.PendingChoices[i]; c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func TestOldB579BloodFunnelDiscountsNoncreatureSpellsAndCountersUnlessYouSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.StockLibraryFor(g, me, 5)
	effectstest.B12Push(g, me.ID, "Blood Funnel", "Enchantment", oldB579Funnel, 0, 0)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", oldB579Bear, 1, 1)
	// Divination costs {2}{U}; with {2} less it costs {U}.
	div := effectstest.HandCardFull(me, "Divination", "Sorcery", "{2}{U}", oldB579Divination, []string{"U"})
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.CastSpell(me.ID, div, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast Divination for {U}: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	prompt := oldB579PayPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{fodder}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d: the spell resolved and drew two", got, hand+2)
	}

	// Declined: the spell is countered.
	div2 := effectstest.HandCardFull(me, "Divination", "Sorcery", "{2}{U}", oldB579Divination, []string{"U"})
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.CastSpell(me.ID, div2, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("second cast: %v", err)
	}
	hand = me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d: the countered spell drew nothing", got, hand)
	}
	if z := effectstest.B12ZoneOf(g, div2); z != game.ZoneGraveyard {
		t.Errorf("the countered spell is in %s, want the graveyard", z)
	}
}

func TestOldB579BloodFunnelCountersWhenYouHaveNoCreatureToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.StockLibraryFor(g, me, 5)
	effectstest.B12Push(g, me.ID, "Blood Funnel", "Enchantment", oldB579Funnel, 0, 0)
	div := effectstest.HandCardFull(me, "Divination", "Sorcery", "{2}{U}", oldB579Divination, []string{"U"})
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.CastSpell(me.ID, div, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if p := oldB579PayPrompt(g, me.ID); p != nil {
		effectstest.AnswerPayUnless(t, g, me.ID, true)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d: with no creature the spell is countered", got, hand)
	}
}

func TestOldB579LightOfSanctionPreventsDamageFromYourSourcesToYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Light of Sanction", "Enchantment", oldB579Light, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB579Bear, 5, 5)
	friend := effectstest.B12Creature(g, me.ID, "Friend", oldB579Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB579Bear, 5, 5)
	effectstest.B27Damage(g, friend, mine, 3)
	if got := effectstest.Pr6Marked(g, mine); got != 0 {
		t.Errorf("a source you control dealt %d to your creature, want 0", got)
	}
	effectstest.B27Damage(g, theirs, mine, 3)
	if got := effectstest.Pr6Marked(g, mine); got != 3 {
		t.Errorf("an opponent's source dealt %d to your creature, want 3", got)
	}
	effectstest.B27Damage(g, mine, theirs, 3)
	if got := effectstest.Pr6Marked(g, theirs); got != 3 {
		t.Errorf("your source dealt %d to their creature, want 3", got)
	}
	life := effectstest.LifeOf(g, me.ID)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(friend, me.ID, 2) })
	if got := effectstest.LifeOf(g, me.ID); got != life-2 {
		t.Errorf("damage to you from your source: life %d, want %d (a player is not a creature)", got, life-2)
	}
}

func TestOldB579MindmoilPutsYourHandOnTheBottomAndDrawsThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.StockLibraryFor(g, me, 12)
	effectstest.B12Push(g, me.ID, "Mindmoil", "Enchantment", oldB579Mindmoil, 0, 0)
	var held []uuid.UUID
	for _, n := range []string{"Held A", "Held B"} {
		held = append(held, effectstest.HandSpell(me, n, "Sorcery", ""))
	}
	spell := effectstest.HandCardFull(me, "Shock", "Instant", "{R}", "", []string{"R"})
	wantHand := me.Hand.Size() - 1
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.PlayerRefs(g.Seats[1].ID)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no order prompt for the two held cards")
	}
	if len(order.ScryCards) != wantHand {
		t.Fatalf("%d cards to order, want the whole hand (%d)", len(order.ScryCards), wantHand)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, order.ScryCards, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != wantHand {
		t.Errorf("hand %d, want %d (every card put on the bottom was replaced by a draw)", me.Hand.Size(), wantHand)
	}
	for _, id := range held {
		if me.Hand.Contains(id) {
			t.Errorf("held card %s is still in hand", id)
		}
	}
}

func TestOldB579QuickchangeRecolorsTheTargetAndDrawsACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.StockLibraryFor(g, me, 5)
	bear := effectstest.B27Push(g, me.ID, "Green Bear", oldB579Bear, "", "{1}{G}", 2, 2, "G")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Quickchange", "Instant", oldB579Quick, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.PendingOfKind(g, game.PendingChoiceColor); c == nil || !c.ColorMulti {
		t.Fatalf("prompt %+v, want a multi color pick", c)
	}
	effectstest.AnswerColor(t, g, me.ID, "UR")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB9Colors(t, g, bear); got != "UR" && got != "RU" {
		t.Errorf("colors %q, want blue and red only", got)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d (one card drawn after the choice)", got, hand+1)
	}
	effectstest.RunCleanup(t, g)
	if got := effectstest.TmtB9Colors(t, g, bear); got != "G" {
		t.Errorf("after the turn colors %q, want green again", got)
	}
}

func TestOldB579TwistedJusticeMakesTheTargetSacrificeAndYouDrawItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.StockLibraryFor(g, me, 8)
	effectstest.StockLibraryFor(g, opp, 8)
	small := effectstest.B12Creature(g, opp.ID, "Small", oldB579Bear, 1, 1)
	big := effectstest.B12Creature(g, opp.ID, "Big", oldB579Bear, 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB579Bear, 6, 6)
	hand, oppHand := me.Hand.Size(), opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Twisted Justice", "Sorcery", oldB579Justice, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || !g.Battlefield.Contains(small) || !g.Battlefield.Contains(mine) {
		t.Error("the target did not sacrifice the creature they chose")
	}
	if got := me.Hand.Size(); got != hand+4 {
		t.Errorf("caster hand %d, want %d (the Big one's power, 4)", got, hand+4)
	}
	if got := opp.Hand.Size(); got != oppHand {
		t.Errorf("target hand %d, want %d: the target draws nothing", got, oppHand)
	}
}

func TestOldB579FestivalOfTheGuildpactPreventsTheNextXDamageToYouAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.StockLibraryFor(g, me, 5)
	src := effectstest.B12Creature(g, opp.ID, "Source", oldB579Bear, 1, 1)
	hand := me.Hand.Size()
	spell := effectstest.HandCardFull(me, "Festival of the Guildpact", "Instant", "{X}{W}", oldB579Festival, []string{"W"})
	effectstest.FloatMana(t, g, me, "{W}{C}{C}{C}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 3}); err != nil {
		t.Fatalf("cast with X=3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d: the spell left and one card was drawn", got, hand+1)
	}
	life := effectstest.LifeOf(g, me.ID)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 5) })
	if got := effectstest.LifeOf(g, me.ID); got != life-2 {
		t.Errorf("life %d, want %d: 3 of the 5 damage was prevented", got, life-2)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 2) })
	if got := effectstest.LifeOf(g, me.ID); got != life-4 {
		t.Errorf("life %d, want %d: the shield was spent", got, life-4)
	}
}

func oldB579EyeTable(t *testing.T) (g *game.Game, me, opp *game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	// The Eye belongs to the other player: its trigger is theirs, the offer is the caster's.
	effectstest.B12Push(g, opp.ID, "Eye of the Storm", "Enchantment", oldB579Eye, 0, 0)
	return g, me, opp
}

func TestOldB579EyeOfTheStormExilesTheSpellAndOffersTheCasterACopy(t *testing.T) {
	g, me, opp := oldB579EyeTable(t)
	life := effectstest.LifeOf(g, opp.ID)
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", oldB579Bolt, effectstest.PlayerRefs(opp.ID))
	for i := 0; i < 12 && effectstest.ResolutionCastOfferFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	offer := effectstest.ResolutionCastOfferFor(g, me.ID)
	if offer == nil {
		t.Fatal("the caster was offered no copy")
	}
	if !effectstest.InExile(g, bolt) {
		t.Error("the Lightning Bolt was not exiled")
	}
	if len(offer.DiscardOptions) != 1 {
		t.Errorf("%d copies offered, want 1", len(offer.DiscardOptions))
	}
	if effectstest.ResolutionCastOfferFor(g, opp.ID) != nil {
		t.Error("the Eye's controller was offered the copy")
	}
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life {
		t.Errorf("opponent life %d, want %d: the Bolt never resolved", got, life)
	}
	if n := len(effectstest.CardCopies(g)); n != 0 {
		t.Errorf("%d card copies remain, want none (an uncast copy ceases to exist)", n)
	}
}

func TestOldB579EyeOfTheStormTriggersAgainOnTheCastCopy(t *testing.T) {
	g, me, opp := oldB579EyeTable(t)
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", oldB579Bolt, effectstest.PlayerRefs(opp.ID))
	for i := 0; i < 12 && effectstest.ResolutionCastOfferFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	offer := effectstest.ResolutionCastOfferFor(g, me.ID)
	if offer == nil {
		t.Fatal("no first offer")
	}
	cp := offer.DiscardOptions[0]
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, cp, game.CastSpellParams{Strict: true, Targets: effectstest.PlayerRefs(opp.ID)}); err != nil {
		t.Fatalf("cast the copy: %v", err)
	}
	for i := 0; i < 12 && effectstest.ResolutionCastOfferFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if effectstest.ResolutionCastOfferFor(g, me.ID) == nil {
		t.Fatal("casting the copy did not offer another copy: Eye did not trigger again")
	}
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life {
		t.Errorf("opponent life %d, want %d: Eye exiled the cast copy before it resolved", got, life)
	}
}

func TestOldB579EyeOfTheStormIgnoresACreatureSpell(t *testing.T) {
	g, me, _ := oldB579EyeTable(t)
	bear := effectstest.HandCardFull(me, "Grizzly Bears", oldB579Bear, "{1}{G}", "", []string{"G"})
	effectstest.FloatMana(t, g, me, "{G}{G}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil || effectstest.InExile(g, bear) {
		t.Error("a creature spell triggered the Eye")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature did not resolve")
	}
}
