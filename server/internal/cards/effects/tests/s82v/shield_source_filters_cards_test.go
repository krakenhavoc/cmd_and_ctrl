package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shield_source_filters_cards_test.go — #2026 (ADR 0108 §7, amendment of
// 2026-10-05): every card the source filter shipped, against the damage
// it must stop, the damage it must not, and a source that changes after
// the shield resolved (CR 609.7b: the filter is read as the damage would
// be dealt).

const (
	sfAlabarasCarpet      = "0f5b0c77-1e3d-46a1-ae0e-03ed79196cd9"
	sfArachnogenesis      = "b655bee5-52d3-467e-b16c-cfc2edf2b1a1"
	sfBenalishMissionary  = "84fdcfd3-2b22-4570-af34-7e3f55f97466"
	sfComeuppance         = "01f29158-2dc9-4b9d-b726-add5d3fd5782"
	sfDeepWood            = "3f01f627-9fbd-470b-8001-974784ccf421"
	sfFogOfWar            = "103ca069-0bfe-4976-bee1-75406273875d"
	sfFrontlineStrategist = "6e202114-48f3-46d2-bc54-df766d149d9d"
	sfGaladhrimAmbush     = "3b22d21f-e19a-40df-840b-a2b7f11f79c2"
	sfHarmlessAssault     = "bfcc3a16-1ca8-4112-a10e-d4e52d3daa8d"
	sfHazeFrog            = "d4cbf16d-6da8-4018-8398-8e263ef6c69e"
	sfHeavyFog            = "a2006755-8812-4aad-8567-e8df6e8923da"
	sfHindervines         = "8d13fabc-1e0f-41f2-8873-9f44b68f7e43"
	sfHuntersAmbush       = "1f085891-0d99-46a6-8f09-b84f44d701f8"
	sfInspireAwe          = "9ca539cd-9876-4c13-b220-d553c17f2378"
	sfJudgment            = "eb226c0d-a71b-4f9c-9fce-47d61dcfa8f8"
	sfLithomancersFocus   = "a8a27a87-fc4c-44b5-9ba8-505bbf42164e"
	sfMoonmist            = "2c306e87-e3c8-4066-9be5-570e2f2d6bad"
	sfObscuringHaze       = "d0a31db0-69db-405c-99a3-945617900c54"
	sfRepel               = "ed198c68-c219-469f-b133-737b060cfbfc"
	sfTanglesap           = "a533df83-782f-4f77-a0be-312ae56f6447"
	sfTerrifyingPresence  = "ca8f6e19-6ef3-4d77-8312-867767ffeda6"
	sfThwartTheEnemy      = "5799baed-3457-4bf2-adf3-239a41dcc1c8"
	sfUndergrowth         = "f125b6b0-c5ff-44d1-bf99-ebc537d37fc3"
	sfVineSnare           = "f3303d45-b3ee-4389-8172-d9ffa8bf251b"
	sfWindsOfQalSisma     = "e7871b4d-a408-4377-beee-6b1d3c7dd57d"
)

// sfCast has `p` cast a card where the game stands — in combat, with
// whoever's priority — and resolves the stack.
func sfCast(t *testing.T, g *game.Game, p *game.Player, card game.Card, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	card.InstanceID = uuid.New()
	card.Owner, card.Controller = p.ID, p.ID
	p.Hand.PushTop(card)
	if err := g.CastSpell(p.ID, card.InstanceID, params); err != nil {
		t.Fatalf("CastSpell %s: %v", card.Name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return card.InstanceID
}

// sfAttackHold declares the attacks and locks them in, leaving the
// cursor in the declare attackers step: the window in which a defender
// casts Heavy Fog.
func sfAttackHold(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	// CR 508.2: the declaration is locked in as a player is next about
	// to receive priority (the state checks).
	g.RunStateChecksForTest()
	if g.Turn.Step != game.StepDeclareAttackers {
		t.Fatalf("left the declare attackers step (at %v)", g.Turn.Step)
	}
}

func sfInstant(name, oracle string) game.Card {
	return game.Card{Name: name, TypeLine: "Instant", OracleID: oracle}
}

// Fog of War, Vine Snare and Hindervines: power and counters, read as
// the creature deals its combat damage.
func TestShieldFilterCardsPowerAndCounters(t *testing.T) {
	t.Run("Fog of War", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		small := effectstest.SfPush(g, me.ID, "Small", "Creature — Test", 3, []string{"G"})
		big := effectstest.SfPush(g, me.ID, "Big", "Creature — Test", 4, []string{"G"})
		grown := effectstest.SfPush(g, me.ID, "Grown", "Creature — Test", 2, []string{"G"})
		effectstest.Pr7bAttack(t, g, opp.ID, small, big, grown)
		life := me.Life
		sfCast(t, g, me, sfInstant("Fog of War", sfFogOfWar), game.CastSpellParams{})
		if me.Life != life+3 {
			t.Fatalf("life %d, want %d: 1 for each of three creatures", me.Life, life+3)
		}
		effectstest.FindBattlefieldCardForTest(g, grown).Counters = map[string]int{game.CounterPlusOne: 2}
		if lost := effectstest.SfCombat(t, g)[opp.ID]; lost != 8 {
			t.Fatalf("defender lost %d, want 8: the 4-power creature and the one grown to 4 deal theirs", lost)
		}
	})
	t.Run("Vine Snare", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		four := effectstest.SfPush(g, me.ID, "Four", "Creature — Test", 4, []string{"G"})
		five := effectstest.SfPush(g, me.ID, "Five", "Creature — Test", 5, []string{"G"})
		effectstest.Pr7bAttack(t, g, opp.ID, four, five)
		sfCast(t, g, me, sfInstant("Vine Snare", sfVineSnare), game.CastSpellParams{})
		if lost := effectstest.SfCombat(t, g)[opp.ID]; lost != 5 {
			t.Fatalf("defender lost %d, want 5", lost)
		}
	})
	t.Run("Hindervines", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		plain := effectstest.SfPush(g, me.ID, "Plain", "Creature — Test", 2, []string{"G"})
		late := effectstest.SfPush(g, me.ID, "Late", "Creature — Test", 3, []string{"G"})
		effectstest.Pr7bAttack(t, g, opp.ID, plain, late)
		sfCast(t, g, me, sfInstant("Hindervines", sfHindervines), game.CastSpellParams{})
		effectstest.FindBattlefieldCardForTest(g, late).Counters = map[string]int{game.CounterPlusOne: 1}
		if lost := effectstest.SfCombat(t, g)[opp.ID]; lost != 4 {
			t.Fatalf("defender lost %d, want 4: only the creature given a counter after the spell deals damage", lost)
		}
	})
}

// "Non-", "nongreen", "without trample", "other than Werewolves and
// Wolves", Inspire Awe's "except" and Undergrowth's "doesn't affect".
func TestShieldFilterCardsNegations(t *testing.T) {
	t.Run("Tanglesap", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		trampler := effectstest.SfPush(g, me.ID, "Trampler", "Creature — Test", 3, []string{"G"}, "trample")
		plain := effectstest.SfPush(g, me.ID, "Plain", "Creature — Test", 2, []string{"G"})
		blocker := effectstest.SfPush(g, opp.ID, "Blocker", "Creature — Test", 2, []string{"G"})
		effectstest.Pr7bAttack(t, g, opp.ID, trampler, plain)
		effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blocker: plain})
		sfCast(t, g, me, sfInstant("Tanglesap", sfTanglesap), game.CastSpellParams{})
		lost := effectstest.SfCombat(t, g)
		if lost[opp.ID] != 3 || effectstest.Pr6Marked(g, plain) != 0 || effectstest.Pr6Marked(g, blocker) != 0 {
			t.Fatalf("defender lost %d (want 3), plain %d, blocker %d (want 0 each)", lost[opp.ID], effectstest.Pr6Marked(g, plain), effectstest.Pr6Marked(g, blocker))
		}
	})
	t.Run("Hunter's Ambush", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		green := effectstest.SfPush(g, me.ID, "Green", "Creature — Test", 2, []string{"G"})
		red := effectstest.SfPush(g, me.ID, "Red", "Creature — Test", 3, []string{"R"})
		golem := effectstest.SfPush(g, me.ID, "Golem", "Artifact Creature — Golem", 4, nil)
		effectstest.Pr7bAttack(t, g, opp.ID, green, red, golem)
		sfCast(t, g, opp, sfInstant("Hunter's Ambush", sfHuntersAmbush), game.CastSpellParams{})
		if lost := effectstest.SfCombat(t, g)[opp.ID]; lost != 2 {
			t.Fatalf("defender lost %d, want 2: only the green creature deals damage", lost)
		}
	})
	t.Run("Repel the Abominable", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		human := effectstest.SfPush(g, opp.ID, "Human", "Creature — Human Warrior", 2, []string{"W"})
		bear := effectstest.SfPush(g, opp.ID, "Bear", "Creature — Bear", 2, []string{"G"})
		rock := effectstest.SfPush(g, opp.ID, "Rock", "Artifact", 0, nil)
		effectstest.CastCatalogSpell(t, g, "Repel the Abominable", "Instant", sfRepel, nil)
		effectstest.PassPriorityAroundTable(t, g)
		life := me.Life
		effectstest.Pr7Hit(t, g, human, me.ID, 2)
		effectstest.Pr7Hit(t, g, bear, me.ID, 2)
		effectstest.Pr7Hit(t, g, rock, me.ID, 2)
		if me.Life != life-2 {
			t.Fatalf("life %d, want %d: only the Human's damage is dealt", me.Life, life-2)
		}
	})
	t.Run("Moonmist", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		wolf := effectstest.SfPush(g, me.ID, "Wolf", "Creature — Wolf", 2, []string{"G"})
		bear := effectstest.SfPush(g, me.ID, "Bear", "Creature — Bear", 3, []string{"G"})
		effectstest.Pr7bAttack(t, g, opp.ID, wolf, bear)
		sfCast(t, g, opp, sfInstant("Moonmist", sfMoonmist), game.CastSpellParams{})
		if lost := effectstest.SfCombat(t, g)[opp.ID]; lost != 2 {
			t.Fatalf("defender lost %d, want 2: only the Wolf deals damage", lost)
		}
	})
	t.Run("Arachnogenesis", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := effectstest.SfPush(g, me.ID, "A", "Creature — Test", 3, []string{"R"})
		b := effectstest.SfPush(g, me.ID, "B", "Creature — Test", 2, []string{"R"})
		sfAttackHold(t, g, opp.ID, a, b)
		sfCast(t, g, opp, sfInstant("Arachnogenesis", sfArachnogenesis), game.CastSpellParams{})
		if n := effectstest.OnBattlefieldNamed(g, "Spider"); n != 2 {
			t.Fatalf("%d Spiders, want 2: one for each creature attacking the caster", n)
		}
		var spider uuid.UUID
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Spider" {
				spider = c.InstanceID
			}
		}
		effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{spider: a})
		lost := effectstest.SfCombat(t, g)
		if lost[opp.ID] != 0 || effectstest.Pr6Marked(g, a) != 1 {
			t.Fatalf("defender lost %d (want 0), attacker marked %d (want the Spider's 1)", lost[opp.ID], effectstest.Pr6Marked(g, a))
		}
	})
	t.Run("Galadhrim Ambush", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := effectstest.SfPush(g, me.ID, "A", "Creature — Test", 3, []string{"R"})
		elf := effectstest.SfPush(g, me.ID, "Elf", "Creature — Elf Druid", 2, []string{"G"})
		effectstest.Pr7bAttack(t, g, opp.ID, a, elf)
		sfCast(t, g, opp, sfInstant("Galadhrim Ambush", sfGaladhrimAmbush), game.CastSpellParams{})
		if n := effectstest.OnBattlefieldNamed(g, "Elf Warrior"); n != 2 {
			t.Fatalf("%d Elf Warriors, want 2: one for each attacking creature", n)
		}
		if lost := effectstest.SfCombat(t, g)[opp.ID]; lost != 2 {
			t.Fatalf("defender lost %d, want 2: only the Elf deals damage", lost)
		}
	})
	t.Run("Inspire Awe", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		worn := effectstest.SfPush(g, me.ID, "Worn", "Creature — Test", 2, []string{"G"})
		spirit := effectstest.SfPush(g, me.ID, "Spirit", "Enchantment Creature — Spirit", 3, []string{"W"})
		plain := effectstest.SfPush(g, me.ID, "Plain", "Creature — Test", 4, []string{"G"})
		effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Aura", TypeLine: "Enchantment — Aura",
			AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: worn}})
		effectstest.Pr7bAttack(t, g, opp.ID, worn, spirit, plain)
		sfCast(t, g, opp, sfInstant("Inspire Awe", sfInspireAwe), game.CastSpellParams{})
		effectstest.AnswerScryKeepAll(t, g, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if lost := effectstest.SfCombat(t, g)[opp.ID]; lost != 5 {
			t.Fatalf("defender lost %d, want 5: the enchanted creature and the enchantment creature deal theirs", lost)
		}
	})
	t.Run("Undergrowth", func(t *testing.T) {
		for _, paid := range []bool{false, true} {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			red := effectstest.SfPush(g, me.ID, "Red", "Creature — Test", 3, []string{"R"})
			green := effectstest.SfPush(g, me.ID, "Green", "Creature — Test", 2, []string{"G"})
			effectstest.Pr7bAttack(t, g, opp.ID, red, green)
			params := game.CastSpellParams{}
			want := 0
			if paid {
				params.OptionalCosts = []int{0}
				want = 3
			}
			sfCast(t, g, opp, sfInstant("Undergrowth", sfUndergrowth), params)
			if lost := effectstest.SfCombat(t, g)[opp.ID]; lost != want {
				t.Fatalf("paid %v: defender lost %d, want %d", paid, lost, want)
			}
		}
	})
}

// "Attacking creatures", "attacking creatures without flying" and
// "target blocked creature".
func TestShieldFilterCardsCombatStatus(t *testing.T) {
	t.Run("Harmless Assault", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		att := effectstest.SfPush(g, me.ID, "Attacker", "Creature — Test", 3, []string{"R"})
		blk := effectstest.SfPush(g, opp.ID, "Blocker", "Creature — Test", 2, []string{"G"})
		effectstest.Pr7bAttack(t, g, opp.ID, att)
		effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blk: att})
		sfCast(t, g, opp, sfInstant("Harmless Assault", sfHarmlessAssault), game.CastSpellParams{})
		effectstest.SfCombat(t, g)
		if effectstest.Pr6Marked(g, blk) != 0 || effectstest.Pr6Marked(g, att) != 2 {
			t.Fatalf("blocker %d (want 0), attacker %d (want the blocker's 2)", effectstest.Pr6Marked(g, blk), effectstest.Pr6Marked(g, att))
		}
	})
	t.Run("Heavy Fog", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if err := effectstest.CastCatalogSpellErr(t, g, "Heavy Fog", "Instant", sfHeavyFog, nil); err == nil {
			t.Fatal("Heavy Fog was cast outside the declare attackers step")
		}
		att := effectstest.SfPush(g, me.ID, "Attacker", "Creature — Test", 3, []string{"R"})
		blocked := effectstest.SfPush(g, me.ID, "Blocked", "Creature — Test", 2, []string{"R"})
		home := effectstest.SfPush(g, me.ID, "Home", "Creature — Test", 1, []string{"R"})
		blk := effectstest.SfPush(g, opp.ID, "Blocker", "Creature — Test", 2, []string{"G"})
		sfAttackHold(t, g, opp.ID, att, blocked)
		sfCast(t, g, opp, sfInstant("Heavy Fog", sfHeavyFog), game.CastSpellParams{})
		life := opp.Life
		effectstest.Pr7Hit(t, g, att, opp.ID, 2)
		effectstest.Pr7Hit(t, g, home, opp.ID, 1)
		if opp.Life != life-1 {
			t.Fatalf("life %d, want %d: the attacker's noncombat damage prevented, the other creature's dealt", opp.Life, life-1)
		}
		effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blk: blocked})
		lost := effectstest.SfCombat(t, g)
		if lost[opp.ID] != 0 || effectstest.Pr6Marked(g, blk) != 2 {
			t.Fatalf("defender lost %d (want 0), blocker marked %d (want 2: only damage to you is prevented)", lost[opp.ID], effectstest.Pr6Marked(g, blk))
		}
	})
	t.Run("Deep Wood needs you to be attacked", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
		att := effectstest.SfPush(g, me.ID, "Attacker", "Creature — Test", 3, []string{"R"})
		sfAttackHold(t, g, third.ID, att)
		card := sfInstant("Deep Wood", sfDeepWood)
		card.InstanceID, card.Owner, card.Controller = uuid.New(), opp.ID, opp.ID
		opp.Hand.PushTop(card)
		if err := g.CastSpell(opp.ID, card.InstanceID, game.CastSpellParams{}); err == nil {
			t.Fatal("Deep Wood was cast by a player nobody attacked")
		}
		sfCast(t, g, third, sfInstant("Deep Wood", sfDeepWood), game.CastSpellParams{})
		if lost := effectstest.SfCombat(t, g)[third.ID]; lost != 0 {
			t.Fatalf("attacked player lost %d, want 0", lost)
		}
	})
	t.Run("Al-abara's Carpet", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		carpet := effectstest.PushCatalogPermanent(g, opp.ID, "Al-abara's Carpet", "Artifact", sfAlabarasCarpet, false)
		flier := effectstest.SfPush(g, me.ID, "Flier", "Creature — Test", 3, []string{"U"}, "flying")
		walker := effectstest.SfPush(g, me.ID, "Walker", "Creature — Test", 2, []string{"G"})
		effectstest.Pr7bAttack(t, g, opp.ID, flier, walker)
		effectstest.Pr7Activate(t, g, opp.ID, carpet, 0, game.ActivateAbilityParams{})
		if lost := effectstest.SfCombat(t, g)[opp.ID]; lost != 3 {
			t.Fatalf("defender lost %d, want 3: only the flier deals damage", lost)
		}
	})
	t.Run("Benalish Missionary", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		missionary := effectstest.PushCatalogPermanent(g, opp.ID, "Benalish Missionary", "Creature — Human Cleric", sfBenalishMissionary, false)
		blocked := effectstest.SfPush(g, me.ID, "Blocked", "Creature — Test", 3, []string{"R"})
		free := effectstest.SfPush(g, me.ID, "Free", "Creature — Test", 2, []string{"R"})
		blk := effectstest.SfPush(g, opp.ID, "Blocker", "Creature — Test", 1, []string{"G"})
		effectstest.Pr7bAttack(t, g, opp.ID, blocked, free)
		effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blk: blocked})
		if err := g.ActivateCatalogAbility(opp.ID, missionary, 0, game.ActivateAbilityParams{Targets: effectstest.Pr7bTargets(free)}); err == nil {
			t.Fatal("an unblocked creature was a legal target for \"target blocked creature\"")
		}
		effectstest.Pr7Activate(t, g, opp.ID, missionary, 0, game.ActivateAbilityParams{Targets: effectstest.Pr7bTargets(blocked)})
		lost := effectstest.SfCombat(t, g)
		if lost[opp.ID] != 2 || effectstest.Pr6Marked(g, blk) != 0 {
			t.Fatalf("defender lost %d (want the free attacker's 2), blocker %d (want 0)", lost[opp.ID], effectstest.Pr6Marked(g, blk))
		}
	})
}

// "Your opponents control", "you don't control", and the per-source
// follow-ups of Comeuppance and Judgment of Alexander.
func TestShieldFilterCardsController(t *testing.T) {
	t.Run("Thwart the Enemy", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.SfPush(g, me.ID, "Mine", "Creature — Test", 2, []string{"G"})
		theirs := effectstest.SfPush(g, opp.ID, "Theirs", "Creature — Test", 2, []string{"R"})
		effectstest.CastCatalogSpell(t, g, "Thwart the Enemy", "Instant", sfThwartTheEnemy, nil)
		effectstest.PassPriorityAroundTable(t, g)
		life := me.Life
		effectstest.Pr7Hit(t, g, theirs, me.ID, 3)
		effectstest.Pr7Hit(t, g, mine, me.ID, 1)
		if me.Life != life-1 {
			t.Fatalf("life %d, want %d: the opponent's creature is stopped, mine is not", me.Life, life-1)
		}
	})
	t.Run("Obscuring Haze", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Commander", TypeLine: "Legendary Creature — Test", Power: 2, Toughness: 2, IsCommander: true})
		theirs := effectstest.SfPush(g, opp.ID, "Theirs", "Creature — Test", 2, []string{"R"})
		effectstest.CastWithAltCost(t, g, "Obscuring Haze", "Instant", sfObscuringHaze, "free")
		effectstest.PassPriorityAroundTable(t, g)
		life := me.Life
		effectstest.Pr7Hit(t, g, theirs, me.ID, 3)
		if me.Life != life {
			t.Fatalf("life %d, want %d", me.Life, life)
		}
	})
	t.Run("Winds of Qal Sisma", func(t *testing.T) {
		for _, ferocious := range []bool{false, true} {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			power := 3
			if ferocious {
				power = 4
			}
			att := effectstest.SfPush(g, me.ID, "Attacker", "Creature — Test", power, []string{"G"})
			blk := effectstest.SfPush(g, opp.ID, "Blocker", "Creature — Test", 2, []string{"R"})
			effectstest.Pr7bAttack(t, g, opp.ID, att)
			effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blk: att})
			sfCast(t, g, me, sfInstant("Winds of Qal Sisma", sfWindsOfQalSisma), game.CastSpellParams{})
			effectstest.SfCombat(t, g)
			wantBlk := 0
			if ferocious {
				wantBlk = 4
			}
			if effectstest.Pr6Marked(g, blk) != wantBlk || effectstest.Pr6Marked(g, att) != 0 {
				t.Fatalf("ferocious %v: blocker %d (want %d), attacker %d (want 0)", ferocious, effectstest.Pr6Marked(g, blk), wantBlk, effectstest.Pr6Marked(g, att))
			}
		}
	})
	t.Run("Comeuppance", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := effectstest.SfPush(g, me.ID, "A", "Creature — Test", 3, []string{"R"})
		b := effectstest.SfPush(g, me.ID, "B", "Creature — Test", 2, []string{"R"})
		rock := effectstest.SfPush(g, me.ID, "Rock", "Artifact", 0, nil)
		effectstest.Pr7bAttack(t, g, opp.ID, a, b)
		sfCast(t, g, opp, sfInstant("Comeuppance", sfComeuppance), game.CastSpellParams{})
		lost := effectstest.SfCombat(t, g)
		if lost[opp.ID] != 0 || effectstest.Pr6Marked(g, a) != 3 || effectstest.Pr6Marked(g, b) != 2 {
			t.Fatalf("defender lost %d (want 0); A %d (want 3), B %d (want 2): each creature is dealt its own damage",
				lost[opp.ID], effectstest.Pr6Marked(g, a), effectstest.Pr6Marked(g, b))
		}
		life := me.Life
		effectstest.Pr7Hit(t, g, rock, opp.ID, 2)
		if me.Life != life-2 {
			t.Fatalf("source's controller life %d, want %d: a noncreature source's damage goes to its controller", me.Life, life-2)
		}
	})
	t.Run("Judgment of Alexander", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Commander", TypeLine: "Legendary Creature — Test", Power: 4, Toughness: 4, IsCommander: true})
		a := effectstest.SfPush(g, me.ID, "A", "Creature — Test", 3, []string{"R"})
		b := effectstest.SfPush(g, me.ID, "B", "Creature — Test", 2, []string{"R"})
		effectstest.Pr7bAttack(t, g, opp.ID, a, b)
		sfCast(t, g, opp, sfInstant("Judgment of Alexander", sfJudgment), game.CastSpellParams{})
		lost := effectstest.SfCombat(t, g)
		if lost[opp.ID] != 0 || effectstest.Pr6Marked(g, a) != 4 || effectstest.Pr6Marked(g, b) != 4 {
			t.Fatalf("defender lost %d (want 0); A %d, B %d (want the commander's 4 each)", lost[opp.ID], effectstest.Pr6Marked(g, a), effectstest.Pr6Marked(g, b))
		}
	})
}

// Objects pinned as the shield resolves: Terrifying Presence's target,
// Haze Frog itself; Lithomancer's Focus's colourless sources and
// Frontline Strategist's non-Soldiers.
func TestShieldFilterCardsPinnedAndProtected(t *testing.T) {
	t.Run("Terrifying Presence", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		chosen := effectstest.SfPush(g, me.ID, "Chosen", "Creature — Test", 3, []string{"G"})
		other := effectstest.SfPush(g, me.ID, "Other", "Creature — Test", 2, []string{"G"})
		effectstest.Pr7bAttack(t, g, opp.ID, chosen, other)
		sfCast(t, g, me, game.Card{Name: "Terrifying Presence", TypeLine: "Instant", OracleID: sfTerrifyingPresence},
			game.CastSpellParams{Targets: effectstest.Pr7bTargets(chosen)})
		if lost := effectstest.SfCombat(t, g)[opp.ID]; lost != 3 {
			t.Fatalf("defender lost %d, want 3: only the target deals damage", lost)
		}
	})
	t.Run("Haze Frog", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := effectstest.SfPush(g, me.ID, "A", "Creature — Test", 3, []string{"R"})
		b := effectstest.SfPush(g, me.ID, "B", "Creature — Test", 2, []string{"R"})
		sfAttackHold(t, g, opp.ID, a, b)
		frog := sfCast(t, g, opp, game.Card{Name: "Haze Frog", TypeLine: "Creature — Frog", OracleID: sfHazeFrog,
			Power: 2, Toughness: 1}, game.CastSpellParams{})
		effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{frog: a})
		lost := effectstest.SfCombat(t, g)
		if lost[opp.ID] != 0 || effectstest.Pr6Marked(g, a) != 2 || effectstest.Pr6Marked(g, frog) != 0 {
			t.Fatalf("defender lost %d (want 0); A %d (want the Frog's 2), Frog %d (want 0)", lost[opp.ID], effectstest.Pr6Marked(g, a), effectstest.Pr6Marked(g, frog))
		}
	})
	t.Run("Lithomancer's Focus", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.SfPush(g, me.ID, "Mine", "Creature — Test", 1, []string{"W"})
		bystander := effectstest.SfPush(g, me.ID, "Bystander", "Creature — Test", 1, []string{"W"})
		golem := effectstest.SfPush(g, opp.ID, "Golem", "Artifact Creature — Golem", 2, nil)
		red := effectstest.SfPush(g, opp.ID, "Red", "Creature — Test", 2, []string{"R"})
		effectstest.CastCatalogSpell(t, g, "Lithomancer's Focus", "Instant", sfLithomancersFocus, effectstest.Pr7bTargets(mine))
		effectstest.PassPriorityAroundTable(t, g)
		if p := effectstest.FindBattlefieldCardForTest(g, mine).CurrentPower(); p != 3 {
			t.Fatalf("power %d, want 3", p)
		}
		effectstest.Pr6Damage(t, g, golem, mine, 2)
		effectstest.Pr6Damage(t, g, red, mine, 1)
		effectstest.Pr6Damage(t, g, golem, bystander, 1)
		if effectstest.Pr6Marked(g, mine) != 1 || effectstest.Pr6Marked(g, bystander) != 1 {
			t.Fatalf("target %d (want the red source's 1), bystander %d (want 1)", effectstest.Pr6Marked(g, mine), effectstest.Pr6Marked(g, bystander))
		}
	})
	t.Run("Frontline Strategist", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		id := effectstest.CastWithAltCost(t, g, "Frontline Strategist", "Creature — Human Soldier", sfFrontlineStrategist, "morph")
		effectstest.PassPriorityAroundTable(t, g)
		soldier := effectstest.SfPush(g, me.ID, "Soldier", "Creature — Human Soldier", 3, []string{"W"})
		bear := effectstest.SfPush(g, me.ID, "Bear", "Creature — Bear", 2, []string{"G"})
		g.WithWriteLock(func() { me.ManaPool.AddMana(effectstest.ManaTokens("W")...) })
		if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
			t.Fatalf("turn face up: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.Pr7bAttack(t, g, opp.ID, soldier, bear)
		if lost := effectstest.SfCombat(t, g)[opp.ID]; lost != 3 {
			t.Fatalf("defender lost %d, want 3: only the Soldier deals damage", lost)
		}
	})
}
