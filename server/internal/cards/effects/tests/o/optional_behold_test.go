package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func castThunderclap(t *testing.T, g *game.Game, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	return effectstest.CastWithTapParams(t, g, "Hulk's Thunderclap", "Sorcery", effectstest.ThunderclapManaCost, effectstest.ThunderclapOracle, params)
}

func castAtomicVision(t *testing.T, g *game.Game, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	return effectstest.CastWithTapParams(t, g, "Hyperion's Atomic Vision", "Instant", effectstest.AtomicVisionManaCost, effectstest.AtomicVisionOracle, params)
}

// beholdBoard is the table the Thunderclap tests share.
type beholdBoard struct {
	g          *game.Game
	me, opp    *game.Player
	mine, foe  uuid.UUID
	gamma      uuid.UUID // a Gamma creature I control
	artifact   uuid.UUID // a noncreature artifact they control
	enchantmnt uuid.UUID
}

func newBeholdBoard(t *testing.T) *beholdBoard {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b := &beholdBoard{g: g, me: me, opp: opp}
	b.mine = effectstest.Pr7Creature(g, me.ID, "Mine", 3, "G")
	b.foe = effectstest.Pr7Creature(g, opp.ID, "Foe", 1, "R")
	b.gamma = effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Gamma Brute", TypeLine: effectstest.GammaTypeLine, Power: 2, Toughness: 2})
	b.artifact = effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Pinger", TypeLine: "Artifact"})
	b.enchantmnt = effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Aura Thing", TypeLine: "Enchantment"})
	return b
}

// Beholding a Gamma you control: the spell is announced with three
// targets, records the paid offer, shows nothing to the table (choosing a
// permanent reveals nothing), and on resolution deals the damage and
// destroys the artifact.
func TestHulksThunderclapBeholdingAPermanent(t *testing.T) {
	b := newBeholdBoard(t)
	id, err := castThunderclap(t, b.g, game.CastSpellParams{
		OptionalCosts: []int{effectstest.BeholdOfferIdx},
		RevealIDs:     []uuid.UUID{b.gamma},
		Targets:       effectstest.Slotted(b.mine, b.foe, b.artifact),
	})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if item := b.g.StackMeta[id]; item == nil || len(item.Paid.OptionalCosts) != 1 || item.Paid.OptionalCosts[0] != effectstest.BeholdOfferIdx {
		t.Fatalf("Paid = %+v, want the behold offer recorded", item)
	}
	if got := effectstest.RevealedEvents(b.g, b.gamma); got != 0 {
		t.Fatalf("choosing a permanent revealed it %d times", got)
	}
	effectstest.PassPriorityAroundTable(t, b.g)
	if got := effectstest.Pr6Marked(b.g, b.foe); got != 3 {
		t.Errorf("the foe took %d, want 3 (Mine's power)", got)
	}
	if effects.OnBattlefield(b.g, b.artifact) {
		t.Error("the beheld cast left the artifact on the battlefield")
	}
	if !effects.OnBattlefield(b.g, b.enchantmnt) {
		t.Error("the enchantment was not a target and must stay")
	}
}

// Beholding a Gamma card in hand reveals it to the table, keeps it in the
// hand, and earns the same bonus.
func TestHulksThunderclapBeholdingFromHand(t *testing.T) {
	b := newBeholdBoard(t)
	gammaCard := effectstest.HandCard(b.me, "Gamma Friend", effectstest.GammaTypeLine)
	if _, err := castThunderclap(t, b.g, game.CastSpellParams{
		OptionalCosts: []int{effectstest.BeholdOfferIdx},
		RevealIDs:     []uuid.UUID{gammaCard},
		Targets:       effectstest.Slotted(b.mine, b.foe, b.enchantmnt),
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if !b.me.Hand.Contains(gammaCard) {
		t.Fatal("revealing moved the card out of the hand")
	}
	if got := effectstest.RevealedEvents(b.g, gammaCard); got != 1 {
		t.Fatalf("%d reveal events for the beheld Gamma, want 1", got)
	}
	effectstest.PassPriorityAroundTable(t, b.g)
	if effects.OnBattlefield(b.g, b.enchantmnt) {
		t.Error("the beheld cast left the enchantment on the battlefield")
	}
	if got := effectstest.Pr6Marked(b.g, b.foe); got != 3 {
		t.Errorf("the foe took %d, want 3", got)
	}
}

// Declined, the spell is the two-target bite: no bonus, and a third
// target is not accepted (CR 601.2c: it exists only for a paid cast).
func TestHulksThunderclapWithoutBeholdHasNoThirdTarget(t *testing.T) {
	b := newBeholdBoard(t)
	if _, err := castThunderclap(t, b.g, game.CastSpellParams{Targets: effectstest.Slotted(b.mine, b.foe, b.artifact)}); err == nil {
		t.Fatal("an unbeholding cast named a third target and was accepted")
	}
	if n := effectstest.StackLen(b.g); n != 0 {
		t.Fatalf("a refused cast left %d object(s) on the stack", n)
	}
	if _, err := castThunderclap(t, b.g, game.CastSpellParams{Targets: effectstest.Slotted(b.mine, b.foe)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, b.g)
	if got := effectstest.Pr6Marked(b.g, b.foe); got != 3 {
		t.Errorf("the foe took %d, want 3", got)
	}
	if !effects.OnBattlefield(b.g, b.artifact) || !effects.OnBattlefield(b.g, b.enchantmnt) {
		t.Error("an unbeholding cast destroyed a noncreature permanent")
	}
}

// The third target is a noncreature artifact or enchantment: a creature,
// an artifact creature and a plain land are refused even when beheld.
func TestHulksThunderclapThirdTargetIsNoncreatureArtifactOrEnchantment(t *testing.T) {
	b := newBeholdBoard(t)
	artCreature := effectstest.ApaPush(b.g, b.opp.ID, b.opp.ID, game.Card{Name: "Golem", TypeLine: "Artifact Creature — Golem", Power: 1, Toughness: 1})
	land := effectstest.ApaPush(b.g, b.opp.ID, b.opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	for _, bad := range []uuid.UUID{b.foe, artCreature, land} {
		if _, err := castThunderclap(t, b.g, game.CastSpellParams{
			OptionalCosts: []int{effectstest.BeholdOfferIdx},
			RevealIDs:     []uuid.UUID{b.gamma},
			Targets:       effectstest.Slotted(b.mine, b.foe, bad),
		}); err == nil {
			t.Fatalf("target %v was accepted for the destroy clause", bad)
		}
	}
	if n := effectstest.StackLen(b.g); n != 0 {
		t.Fatalf("refused casts left %d object(s) on the stack", n)
	}
}

// Behold is refused, and leaves the table as it was, when there is
// nothing to behold, the card named is not a Gamma, the spell itself is
// named, nothing is named, or a card is named for a cast that did not
// take the offer.
func TestHulksThunderclapBeholdRefusals(t *testing.T) {
	b := newBeholdBoard(t)
	nonGamma := effectstest.HandCard(b.me, "Plain Friend", "Creature — Goblin")
	theirs := effectstest.HandCard(b.opp, "Their Gamma", effectstest.GammaTypeLine)
	three := effectstest.Slotted(b.mine, b.foe, b.artifact)
	for name, params := range map[string]game.CastSpellParams{
		"nothing named":                  {OptionalCosts: []int{effectstest.BeholdOfferIdx}, Targets: three},
		"not a Gamma":                    {OptionalCosts: []int{effectstest.BeholdOfferIdx}, RevealIDs: []uuid.UUID{nonGamma}, Targets: three},
		"a Gamma card they own":          {OptionalCosts: []int{effectstest.BeholdOfferIdx}, RevealIDs: []uuid.UUID{theirs}, Targets: three},
		"a creature that is not a Gamma": {OptionalCosts: []int{effectstest.BeholdOfferIdx}, RevealIDs: []uuid.UUID{b.mine}, Targets: three},
		"two cards":                      {OptionalCosts: []int{effectstest.BeholdOfferIdx}, RevealIDs: []uuid.UUID{b.gamma, b.gamma}, Targets: three},
		"a card but no offer":            {RevealIDs: []uuid.UUID{b.gamma}, Targets: effectstest.Slotted(b.mine, b.foe)},
	} {
		if _, err := castThunderclap(t, b.g, params); err == nil {
			t.Errorf("%s: the cast was accepted", name)
		}
		if n := effectstest.StackLen(b.g); n != 0 {
			t.Fatalf("%s: a refused cast left %d object(s) on the stack", name, n)
		}
	}
	// With no Gamma anywhere, the offer cannot be taken at all.
	b.g.WithWriteLock(func() { _ = b.g.DestroyPermanentForEffect(b.gamma, game.DestroyOptions{}) })
	if _, err := castThunderclap(t, b.g, game.CastSpellParams{OptionalCosts: []int{effectstest.BeholdOfferIdx}, Targets: three}); err == nil {
		t.Error("beholding with no Gamma to behold was accepted")
	}
}

// A Gamma the caster does not control is not a behold choice: only the
// caster's permanents and the caster's hand count.
func TestHulksThunderclapCannotBeholdAnOpponentsGamma(t *testing.T) {
	b := newBeholdBoard(t)
	b.g.WithWriteLock(func() { _ = b.g.DestroyPermanentForEffect(b.gamma, game.DestroyOptions{}) })
	theirs := effectstest.ApaPush(b.g, b.opp.ID, b.opp.ID, game.Card{Name: "Their Gamma", TypeLine: effectstest.GammaTypeLine, Power: 1, Toughness: 1})
	if _, err := castThunderclap(t, b.g, game.CastSpellParams{
		OptionalCosts: []int{effectstest.BeholdOfferIdx},
		RevealIDs:     []uuid.UUID{theirs},
		Targets:       effectstest.Slotted(b.mine, b.foe, b.artifact),
	}); err == nil {
		t.Fatal("beholding an opponent's Gamma was accepted")
	}
}

// "Equal to its power" is read as the spell resolves: a counter added in
// response is in the damage (the layer-correct power, CR 608.2h).
func TestHulksThunderclapDamageReadsPowerAtResolution(t *testing.T) {
	b := newBeholdBoard(t)
	big := effectstest.ApaPush(b.g, b.opp.ID, b.opp.ID, game.Card{Name: "Wall of Girth", TypeLine: "Creature — Wall", Power: 0, Toughness: 9})
	if _, err := castThunderclap(t, b.g, game.CastSpellParams{Targets: effectstest.Slotted(b.mine, big)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	b.g.WithWriteLock(func() { _ = b.g.AddCounterForEffect(b.mine, "+1/+1", 2) })
	effectstest.PassPriorityAroundTable(t, b.g)
	if got := effectstest.Pr6Marked(b.g, big); got != 5 {
		t.Errorf("the target took %d, want 5 (3 + two counters)", got)
	}
}

// A source that left in response is an illegal target (CR 608.2b) and
// deals no damage; the destruction is its own sentence and still happens.
func TestHulksThunderclapSourceLeftStillDestroys(t *testing.T) {
	b := newBeholdBoard(t)
	if _, err := castThunderclap(t, b.g, game.CastSpellParams{
		OptionalCosts: []int{effectstest.BeholdOfferIdx},
		RevealIDs:     []uuid.UUID{b.gamma},
		Targets:       effectstest.Slotted(b.mine, b.foe, b.artifact),
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	b.g.WithWriteLock(func() { _ = b.g.DestroyPermanentForEffect(b.mine, game.DestroyOptions{}) })
	effectstest.PassPriorityAroundTable(t, b.g)
	if got := effectstest.Pr6Marked(b.g, b.foe); got != 0 {
		t.Errorf("the foe took %d from a creature that left, want 0", got)
	}
	if effects.OnBattlefield(b.g, b.artifact) {
		t.Error("the destroy clause did not happen")
	}
}

// The beholding cast is a restore point: a snapshot taken with the spell
// on the stack brings the paid offer back, and the destroy still happens.
func TestHulksThunderclapBeholdSurvivesASnapshot(t *testing.T) {
	b := newBeholdBoard(t)
	id, err := castThunderclap(t, b.g, game.CastSpellParams{
		OptionalCosts: []int{effectstest.BeholdOfferIdx},
		RevealIDs:     []uuid.UUID{b.gamma},
		Targets:       effectstest.Slotted(b.mine, b.foe, b.artifact),
	})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	restored := effectstest.RestoreRoundTrip(t, b.g, false)
	if item := restored.StackMeta[id]; item == nil || len(item.Paid.OptionalCosts) != 1 {
		t.Fatalf("the restored item = %+v, want the behold offer recorded", item)
	}
	effectstest.PassPriorityAroundTable(t, restored)
	if effects.OnBattlefield(restored, b.artifact) {
		t.Error("the restored beheld cast did not destroy the artifact")
	}
}

// Undo: a clone taken with the beholding spell on the stack restores the
// paid offer, and the clone does not alias the live record.
func TestHulksThunderclapBeholdSurvivesUndo(t *testing.T) {
	b := newBeholdBoard(t)
	id, err := castThunderclap(t, b.g, game.CastSpellParams{
		OptionalCosts: []int{effectstest.BeholdOfferIdx},
		RevealIDs:     []uuid.UUID{b.gamma},
		Targets:       effectstest.Slotted(b.mine, b.foe, b.artifact),
	})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	mid := b.g.Clone()
	b.g.WithWriteLock(func() { b.g.StackMeta[id].Paid.OptionalCosts = nil })
	b.g.RestoreFrom(mid)
	item := b.g.StackMeta[id]
	if item == nil || item.Paid.OptionalCostTimes(effectstest.BeholdOfferIdx) != 1 {
		t.Fatalf("after undo the item = %+v, want the behold offer paid once", item)
	}
}

// --- Hyperion's Atomic Vision ------------------------------------------

type visionBoard struct {
	g        *game.Game
	me, opp  *game.Player
	hero     uuid.UUID // a Hero I control
	tapped   uuid.UUID // a tapped creature they control
	untapped uuid.UUID
}

func newVisionBoard(t *testing.T) *visionBoard {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b := &visionBoard{g: g, me: me, opp: opp}
	b.hero = effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Hero Dude", TypeLine: effectstest.HeroTypeLine, Power: 2, Toughness: 2})
	b.tapped = effectstest.Pr7Creature(g, opp.ID, "Tapped Foe", 2, "R")
	b.untapped = effectstest.Pr7Creature(g, opp.ID, "Fresh Foe", 2, "R")
	effectstest.FindBattlefieldCardForTest(g, b.tapped).Tapped = true
	return b
}

func scryOpenFor(g *game.Game, chooser uuid.UUID) bool {
	open := false
	g.ReadSnapshot(func() { open = effectstest.ScryChoiceFor(g, chooser) != nil })
	return open
}

// Beheld, a Hero permanent: the tapped creature dies and its caster
// scries 2.
func TestHyperionsAtomicVisionBeheldScries(t *testing.T) {
	b := newVisionBoard(t)
	if _, err := castAtomicVision(t, b.g, game.CastSpellParams{
		OptionalCosts: []int{effectstest.BeholdOfferIdx},
		RevealIDs:     []uuid.UUID{b.hero},
		Targets:       effectstest.Slotted(b.tapped),
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, b.g)
	if effects.OnBattlefield(b.g, b.tapped) {
		t.Error("the tapped creature survived")
	}
	c := effectstest.ScryChoiceFor(b.g, b.me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("scry prompt = %+v, want a scry of 2 for the caster", c)
	}
}

// Beheld from hand: the Hero card is revealed, stays in hand, and the
// scry follows.
func TestHyperionsAtomicVisionBeheldFromHand(t *testing.T) {
	b := newVisionBoard(t)
	heroCard := effectstest.HandCard(b.me, "Hero Friend", effectstest.HeroTypeLine)
	if _, err := castAtomicVision(t, b.g, game.CastSpellParams{
		OptionalCosts: []int{effectstest.BeholdOfferIdx},
		RevealIDs:     []uuid.UUID{heroCard},
		Targets:       effectstest.Slotted(b.tapped),
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if !b.me.Hand.Contains(heroCard) || effectstest.RevealedEvents(b.g, heroCard) != 1 {
		t.Fatalf("hand has it %v, reveal events %d; want kept and revealed once", b.me.Hand.Contains(heroCard), effectstest.RevealedEvents(b.g, heroCard))
	}
	effectstest.PassPriorityAroundTable(t, b.g)
	if !scryOpenFor(b.g, b.me.ID) {
		t.Error("no scry after beholding a Hero card from hand")
	}
}

// Declined: the tapped creature dies and there is no scry.
func TestHyperionsAtomicVisionUnbeheldDoesNotScry(t *testing.T) {
	b := newVisionBoard(t)
	if _, err := castAtomicVision(t, b.g, game.CastSpellParams{Targets: effectstest.Slotted(b.tapped)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, b.g)
	if effects.OnBattlefield(b.g, b.tapped) {
		t.Error("the tapped creature survived")
	}
	if scryOpenFor(b.g, b.me.ID) {
		t.Error("an unbeholding cast scried")
	}
}

// "If a Hero was beheld" is a fact about the announcement (CR 701.4b): the
// Hero leaving in response changes nothing.
func TestHyperionsAtomicVisionBeheldHeroLeavingStillScries(t *testing.T) {
	b := newVisionBoard(t)
	if _, err := castAtomicVision(t, b.g, game.CastSpellParams{
		OptionalCosts: []int{effectstest.BeholdOfferIdx},
		RevealIDs:     []uuid.UUID{b.hero},
		Targets:       effectstest.Slotted(b.tapped),
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	b.g.WithWriteLock(func() { _ = b.g.DestroyPermanentForEffect(b.hero, game.DestroyOptions{}) })
	effectstest.PassPriorityAroundTable(t, b.g)
	if !scryOpenFor(b.g, b.me.ID) {
		t.Error("the Hero left before resolution and the scry was lost")
	}
}

// The target must be tapped; the behold offer needs a Hero.
func TestHyperionsAtomicVisionRefusals(t *testing.T) {
	b := newVisionBoard(t)
	if _, err := castAtomicVision(t, b.g, game.CastSpellParams{Targets: effectstest.Slotted(b.untapped)}); err == nil {
		t.Error("an untapped creature was accepted as the target")
	}
	notHero := effectstest.HandCard(b.me, "Plain Friend", "Creature — Goblin")
	for name, params := range map[string]game.CastSpellParams{
		"nothing named": {OptionalCosts: []int{effectstest.BeholdOfferIdx}, Targets: effectstest.Slotted(b.tapped)},
		"not a Hero":    {OptionalCosts: []int{effectstest.BeholdOfferIdx}, RevealIDs: []uuid.UUID{notHero}, Targets: effectstest.Slotted(b.tapped)},
		"no offer":      {RevealIDs: []uuid.UUID{b.hero}, Targets: effectstest.Slotted(b.tapped)},
	} {
		if _, err := castAtomicVision(t, b.g, params); err == nil {
			t.Errorf("%s: the cast was accepted", name)
		}
	}
	if n := effectstest.StackLen(b.g); n != 0 {
		t.Fatalf("refused casts left %d object(s) on the stack", n)
	}
}

// The beheld scry is a restore point too.
func TestHyperionsAtomicVisionBeheldSurvivesASnapshot(t *testing.T) {
	b := newVisionBoard(t)
	if _, err := castAtomicVision(t, b.g, game.CastSpellParams{
		OptionalCosts: []int{effectstest.BeholdOfferIdx},
		RevealIDs:     []uuid.UUID{b.hero},
		Targets:       effectstest.Slotted(b.tapped),
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	restored := effectstest.RestoreRoundTrip(t, b.g, false)
	effectstest.PassPriorityAroundTable(t, restored)
	if !scryOpenFor(restored, b.me.ID) {
		t.Error("the restored beheld cast did not scry")
	}
}

// --- Register's guards --------------------------------------------------

func TestRegisterRefusesMalformedOptionalBehold(t *testing.T) {
	good := effects.OptionalBehold("a", "Hero")

	// An optional reveal under another key is a mandatory-only shape.
	reveal := good
	reveal.Key = "kicker"
	effectstest.MustPanic(t, "reveals a card", func() {
		effects.CheckEitherCost(effects.Spec{Name: "Kicked Reveal", OptionalCosts: []game.AdditionalCost{reveal}})
	})

	// The behold key with no behold behind it, or with a plain reveal.
	for name, mutate := range map[string]func(*game.AdditionalCost){
		"no reveal":        func(c *game.AdditionalCost) { c.Reveal = nil; c.ManaCost = "{1}" },
		"not a behold":     func(c *game.AdditionalCost) { c.Reveal = &game.RevealCost{Subtype: "Hero"} },
		"no type":          func(c *game.AdditionalCost) { c.Reveal = &game.RevealCost{Behold: true} },
		"a colour":         func(c *game.AdditionalCost) { c.Reveal = &game.RevealCost{Subtype: "Hero", Behold: true, Color: "W"} },
		"mana beside":      func(c *game.AdditionalCost) { c.ManaCost = "{1}" },
		"a repeat":         func(c *game.AdditionalCost) { c.Repeat = 3 },
		"a discard beside": func(c *game.AdditionalCost) { c.DiscardCards = 1 },
	} {
		bad := good
		mutate(&bad)
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: no panic", name)
				}
			}()
			effects.CheckOptionalBehold(effects.Spec{Name: "Bad", OptionalCosts: []game.AdditionalCost{bad}})
		}()
	}

	// Two of them share reveal_ids; so does one beside a mandatory branch.
	effectstest.MustPanic(t, "optional behold costs", func() {
		effects.CheckOptionalBehold(effects.Spec{Name: "Twice", OptionalCosts: []game.AdditionalCost{good, good}})
	})
	effectstest.MustPanic(t, "mandatory reveal", func() {
		effects.CheckOptionalBehold(effects.Spec{Name: "Both", AdditionalCost: effects.BeholdOrPay("a", "Dragon", "{1}"), OptionalCosts: []game.AdditionalCost{good}})
	})

	// The printed shapes pass.
	effects.CheckOptionalBehold(effects.Spec{Name: "Hyperion", OptionalCosts: []game.AdditionalCost{good}})
	effects.CheckOptionalBehold(effects.Spec{Name: "Thunderclap", OptionalCosts: []game.AdditionalCost{effects.WhenPaid(good, effects.TargetCreature("target creature"))}})
	effects.CheckEitherCost(effects.Spec{Name: "Hyperion", OptionalCosts: []game.AdditionalCost{good}})
}
