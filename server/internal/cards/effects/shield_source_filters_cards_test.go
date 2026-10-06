package effects

import (
	"testing"

	"github.com/google/uuid"

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

// sfPush puts a permanent with the given type line, power and colours on
// the battlefield for `owner`, toughness 5 so combat damage marks rather
// than kills.
func sfPush(g *game.Game, owner uuid.UUID, name, typeLine string, power int, colors []string, keywords ...string) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{Name: name, TypeLine: typeLine, Power: power, Toughness: 5,
		Colors: colors, Keywords: keywords})
}

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
	passPriorityAroundTable(t, g)
	return card.InstanceID
}

// sfAttackHold declares the attacks and locks them in, leaving the
// cursor in the declare attackers step: the window in which a defender
// casts Heavy Fog.
func sfAttackHold(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
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

// sfCombat walks to the combat damage step and back out of its priority
// boundary, returning how much life each player lost.
func sfCombat(t *testing.T, g *game.Game) map[uuid.UUID]int {
	t.Helper()
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	pr7bDamageStep(t, g)
	passPriorityAroundTable(t, g)
	lost := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lost[p.ID] = before[p.ID] - p.Life
	}
	return lost
}

// Fog of War, Vine Snare and Hindervines: power and counters, read as
// the creature deals its combat damage.
func TestShieldFilterCardsPowerAndCounters(t *testing.T) {
	t.Run("Fog of War", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		small := sfPush(g, me.ID, "Small", "Creature — Test", 3, []string{"G"})
		big := sfPush(g, me.ID, "Big", "Creature — Test", 4, []string{"G"})
		grown := sfPush(g, me.ID, "Grown", "Creature — Test", 2, []string{"G"})
		pr7bAttack(t, g, opp.ID, small, big, grown)
		life := me.Life
		sfCast(t, g, me, sfInstant("Fog of War", sfFogOfWar), game.CastSpellParams{})
		if me.Life != life+3 {
			t.Fatalf("life %d, want %d: 1 for each of three creatures", me.Life, life+3)
		}
		findBattlefieldCardForTest(g, grown).Counters = map[string]int{game.CounterPlusOne: 2}
		if lost := sfCombat(t, g)[opp.ID]; lost != 8 {
			t.Fatalf("defender lost %d, want 8: the 4-power creature and the one grown to 4 deal theirs", lost)
		}
	})
	t.Run("Vine Snare", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		four := sfPush(g, me.ID, "Four", "Creature — Test", 4, []string{"G"})
		five := sfPush(g, me.ID, "Five", "Creature — Test", 5, []string{"G"})
		pr7bAttack(t, g, opp.ID, four, five)
		sfCast(t, g, me, sfInstant("Vine Snare", sfVineSnare), game.CastSpellParams{})
		if lost := sfCombat(t, g)[opp.ID]; lost != 5 {
			t.Fatalf("defender lost %d, want 5", lost)
		}
	})
	t.Run("Hindervines", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		plain := sfPush(g, me.ID, "Plain", "Creature — Test", 2, []string{"G"})
		late := sfPush(g, me.ID, "Late", "Creature — Test", 3, []string{"G"})
		pr7bAttack(t, g, opp.ID, plain, late)
		sfCast(t, g, me, sfInstant("Hindervines", sfHindervines), game.CastSpellParams{})
		findBattlefieldCardForTest(g, late).Counters = map[string]int{game.CounterPlusOne: 1}
		if lost := sfCombat(t, g)[opp.ID]; lost != 4 {
			t.Fatalf("defender lost %d, want 4: only the creature given a counter after the spell deals damage", lost)
		}
	})
}

// "Non-", "nongreen", "without trample", "other than Werewolves and
// Wolves", Inspire Awe's "except" and Undergrowth's "doesn't affect".
func TestShieldFilterCardsNegations(t *testing.T) {
	t.Run("Tanglesap", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		trampler := sfPush(g, me.ID, "Trampler", "Creature — Test", 3, []string{"G"}, "trample")
		plain := sfPush(g, me.ID, "Plain", "Creature — Test", 2, []string{"G"})
		blocker := sfPush(g, opp.ID, "Blocker", "Creature — Test", 2, []string{"G"})
		pr7bAttack(t, g, opp.ID, trampler, plain)
		pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blocker: plain})
		sfCast(t, g, me, sfInstant("Tanglesap", sfTanglesap), game.CastSpellParams{})
		lost := sfCombat(t, g)
		if lost[opp.ID] != 3 || pr6Marked(g, plain) != 0 || pr6Marked(g, blocker) != 0 {
			t.Fatalf("defender lost %d (want 3), plain %d, blocker %d (want 0 each)", lost[opp.ID], pr6Marked(g, plain), pr6Marked(g, blocker))
		}
	})
	t.Run("Hunter's Ambush", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		green := sfPush(g, me.ID, "Green", "Creature — Test", 2, []string{"G"})
		red := sfPush(g, me.ID, "Red", "Creature — Test", 3, []string{"R"})
		golem := sfPush(g, me.ID, "Golem", "Artifact Creature — Golem", 4, nil)
		pr7bAttack(t, g, opp.ID, green, red, golem)
		sfCast(t, g, opp, sfInstant("Hunter's Ambush", sfHuntersAmbush), game.CastSpellParams{})
		if lost := sfCombat(t, g)[opp.ID]; lost != 2 {
			t.Fatalf("defender lost %d, want 2: only the green creature deals damage", lost)
		}
	})
	t.Run("Repel the Abominable", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		human := sfPush(g, opp.ID, "Human", "Creature — Human Warrior", 2, []string{"W"})
		bear := sfPush(g, opp.ID, "Bear", "Creature — Bear", 2, []string{"G"})
		rock := sfPush(g, opp.ID, "Rock", "Artifact", 0, nil)
		castCatalogSpell(t, g, "Repel the Abominable", "Instant", sfRepel, nil)
		passPriorityAroundTable(t, g)
		life := me.Life
		pr7Hit(t, g, human, me.ID, 2)
		pr7Hit(t, g, bear, me.ID, 2)
		pr7Hit(t, g, rock, me.ID, 2)
		if me.Life != life-2 {
			t.Fatalf("life %d, want %d: only the Human's damage is dealt", me.Life, life-2)
		}
	})
	t.Run("Moonmist", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		wolf := sfPush(g, me.ID, "Wolf", "Creature — Wolf", 2, []string{"G"})
		bear := sfPush(g, me.ID, "Bear", "Creature — Bear", 3, []string{"G"})
		pr7bAttack(t, g, opp.ID, wolf, bear)
		sfCast(t, g, opp, sfInstant("Moonmist", sfMoonmist), game.CastSpellParams{})
		if lost := sfCombat(t, g)[opp.ID]; lost != 2 {
			t.Fatalf("defender lost %d, want 2: only the Wolf deals damage", lost)
		}
	})
	t.Run("Arachnogenesis", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := sfPush(g, me.ID, "A", "Creature — Test", 3, []string{"R"})
		b := sfPush(g, me.ID, "B", "Creature — Test", 2, []string{"R"})
		sfAttackHold(t, g, opp.ID, a, b)
		sfCast(t, g, opp, sfInstant("Arachnogenesis", sfArachnogenesis), game.CastSpellParams{})
		if n := onBattlefieldNamed(g, "Spider"); n != 2 {
			t.Fatalf("%d Spiders, want 2: one for each creature attacking the caster", n)
		}
		var spider uuid.UUID
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Spider" {
				spider = c.InstanceID
			}
		}
		pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{spider: a})
		lost := sfCombat(t, g)
		if lost[opp.ID] != 0 || pr6Marked(g, a) != 1 {
			t.Fatalf("defender lost %d (want 0), attacker marked %d (want the Spider's 1)", lost[opp.ID], pr6Marked(g, a))
		}
	})
	t.Run("Galadhrim Ambush", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := sfPush(g, me.ID, "A", "Creature — Test", 3, []string{"R"})
		elf := sfPush(g, me.ID, "Elf", "Creature — Elf Druid", 2, []string{"G"})
		pr7bAttack(t, g, opp.ID, a, elf)
		sfCast(t, g, opp, sfInstant("Galadhrim Ambush", sfGaladhrimAmbush), game.CastSpellParams{})
		if n := onBattlefieldNamed(g, "Elf Warrior"); n != 2 {
			t.Fatalf("%d Elf Warriors, want 2: one for each attacking creature", n)
		}
		if lost := sfCombat(t, g)[opp.ID]; lost != 2 {
			t.Fatalf("defender lost %d, want 2: only the Elf deals damage", lost)
		}
	})
	t.Run("Inspire Awe", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		worn := sfPush(g, me.ID, "Worn", "Creature — Test", 2, []string{"G"})
		spirit := sfPush(g, me.ID, "Spirit", "Enchantment Creature — Spirit", 3, []string{"W"})
		plain := sfPush(g, me.ID, "Plain", "Creature — Test", 4, []string{"G"})
		apaPush(g, me.ID, me.ID, game.Card{Name: "Aura", TypeLine: "Enchantment — Aura",
			AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: worn}})
		pr7bAttack(t, g, opp.ID, worn, spirit, plain)
		sfCast(t, g, opp, sfInstant("Inspire Awe", sfInspireAwe), game.CastSpellParams{})
		answerScryKeepAll(t, g, opp.ID)
		passPriorityAroundTable(t, g)
		if lost := sfCombat(t, g)[opp.ID]; lost != 5 {
			t.Fatalf("defender lost %d, want 5: the enchanted creature and the enchantment creature deal theirs", lost)
		}
	})
	t.Run("Undergrowth", func(t *testing.T) {
		for _, paid := range []bool{false, true} {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			red := sfPush(g, me.ID, "Red", "Creature — Test", 3, []string{"R"})
			green := sfPush(g, me.ID, "Green", "Creature — Test", 2, []string{"G"})
			pr7bAttack(t, g, opp.ID, red, green)
			params := game.CastSpellParams{}
			want := 0
			if paid {
				params.OptionalCosts = []int{0}
				want = 3
			}
			sfCast(t, g, opp, sfInstant("Undergrowth", sfUndergrowth), params)
			if lost := sfCombat(t, g)[opp.ID]; lost != want {
				t.Fatalf("paid %v: defender lost %d, want %d", paid, lost, want)
			}
		}
	})
}

// "Attacking creatures", "attacking creatures without flying" and
// "target blocked creature".
func TestShieldFilterCardsCombatStatus(t *testing.T) {
	t.Run("Harmless Assault", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		att := sfPush(g, me.ID, "Attacker", "Creature — Test", 3, []string{"R"})
		blk := sfPush(g, opp.ID, "Blocker", "Creature — Test", 2, []string{"G"})
		pr7bAttack(t, g, opp.ID, att)
		pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blk: att})
		sfCast(t, g, opp, sfInstant("Harmless Assault", sfHarmlessAssault), game.CastSpellParams{})
		sfCombat(t, g)
		if pr6Marked(g, blk) != 0 || pr6Marked(g, att) != 2 {
			t.Fatalf("blocker %d (want 0), attacker %d (want the blocker's 2)", pr6Marked(g, blk), pr6Marked(g, att))
		}
	})
	t.Run("Heavy Fog", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if err := castCatalogSpellErr(t, g, "Heavy Fog", "Instant", sfHeavyFog, nil); err == nil {
			t.Fatal("Heavy Fog was cast outside the declare attackers step")
		}
		att := sfPush(g, me.ID, "Attacker", "Creature — Test", 3, []string{"R"})
		blocked := sfPush(g, me.ID, "Blocked", "Creature — Test", 2, []string{"R"})
		home := sfPush(g, me.ID, "Home", "Creature — Test", 1, []string{"R"})
		blk := sfPush(g, opp.ID, "Blocker", "Creature — Test", 2, []string{"G"})
		sfAttackHold(t, g, opp.ID, att, blocked)
		sfCast(t, g, opp, sfInstant("Heavy Fog", sfHeavyFog), game.CastSpellParams{})
		life := opp.Life
		pr7Hit(t, g, att, opp.ID, 2)
		pr7Hit(t, g, home, opp.ID, 1)
		if opp.Life != life-1 {
			t.Fatalf("life %d, want %d: the attacker's noncombat damage prevented, the other creature's dealt", opp.Life, life-1)
		}
		pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blk: blocked})
		lost := sfCombat(t, g)
		if lost[opp.ID] != 0 || pr6Marked(g, blk) != 2 {
			t.Fatalf("defender lost %d (want 0), blocker marked %d (want 2: only damage to you is prevented)", lost[opp.ID], pr6Marked(g, blk))
		}
	})
	t.Run("Deep Wood needs you to be attacked", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
		att := sfPush(g, me.ID, "Attacker", "Creature — Test", 3, []string{"R"})
		sfAttackHold(t, g, third.ID, att)
		card := sfInstant("Deep Wood", sfDeepWood)
		card.InstanceID, card.Owner, card.Controller = uuid.New(), opp.ID, opp.ID
		opp.Hand.PushTop(card)
		if err := g.CastSpell(opp.ID, card.InstanceID, game.CastSpellParams{}); err == nil {
			t.Fatal("Deep Wood was cast by a player nobody attacked")
		}
		sfCast(t, g, third, sfInstant("Deep Wood", sfDeepWood), game.CastSpellParams{})
		if lost := sfCombat(t, g)[third.ID]; lost != 0 {
			t.Fatalf("attacked player lost %d, want 0", lost)
		}
	})
	t.Run("Al-abara's Carpet", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		carpet := pushCatalogPermanent(g, opp.ID, "Al-abara's Carpet", "Artifact", sfAlabarasCarpet, false)
		flier := sfPush(g, me.ID, "Flier", "Creature — Test", 3, []string{"U"}, "flying")
		walker := sfPush(g, me.ID, "Walker", "Creature — Test", 2, []string{"G"})
		pr7bAttack(t, g, opp.ID, flier, walker)
		pr7Activate(t, g, opp.ID, carpet, 0, game.ActivateAbilityParams{})
		if lost := sfCombat(t, g)[opp.ID]; lost != 3 {
			t.Fatalf("defender lost %d, want 3: only the flier deals damage", lost)
		}
	})
	t.Run("Benalish Missionary", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		missionary := pushCatalogPermanent(g, opp.ID, "Benalish Missionary", "Creature — Human Cleric", sfBenalishMissionary, false)
		blocked := sfPush(g, me.ID, "Blocked", "Creature — Test", 3, []string{"R"})
		free := sfPush(g, me.ID, "Free", "Creature — Test", 2, []string{"R"})
		blk := sfPush(g, opp.ID, "Blocker", "Creature — Test", 1, []string{"G"})
		pr7bAttack(t, g, opp.ID, blocked, free)
		pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blk: blocked})
		if err := g.ActivateCatalogAbility(opp.ID, missionary, 0, game.ActivateAbilityParams{Targets: pr7bTargets(free)}); err == nil {
			t.Fatal("an unblocked creature was a legal target for \"target blocked creature\"")
		}
		pr7Activate(t, g, opp.ID, missionary, 0, game.ActivateAbilityParams{Targets: pr7bTargets(blocked)})
		lost := sfCombat(t, g)
		if lost[opp.ID] != 2 || pr6Marked(g, blk) != 0 {
			t.Fatalf("defender lost %d (want the free attacker's 2), blocker %d (want 0)", lost[opp.ID], pr6Marked(g, blk))
		}
	})
}

// "Your opponents control", "you don't control", and the per-source
// follow-ups of Comeuppance and Judgment of Alexander.
func TestShieldFilterCardsController(t *testing.T) {
	t.Run("Thwart the Enemy", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := sfPush(g, me.ID, "Mine", "Creature — Test", 2, []string{"G"})
		theirs := sfPush(g, opp.ID, "Theirs", "Creature — Test", 2, []string{"R"})
		castCatalogSpell(t, g, "Thwart the Enemy", "Instant", sfThwartTheEnemy, nil)
		passPriorityAroundTable(t, g)
		life := me.Life
		pr7Hit(t, g, theirs, me.ID, 3)
		pr7Hit(t, g, mine, me.ID, 1)
		if me.Life != life-1 {
			t.Fatalf("life %d, want %d: the opponent's creature is stopped, mine is not", me.Life, life-1)
		}
	})
	t.Run("Obscuring Haze", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		apaPush(g, me.ID, me.ID, game.Card{Name: "Commander", TypeLine: "Legendary Creature — Test", Power: 2, Toughness: 2, IsCommander: true})
		theirs := sfPush(g, opp.ID, "Theirs", "Creature — Test", 2, []string{"R"})
		castWithAltCost(t, g, "Obscuring Haze", "Instant", sfObscuringHaze, "free")
		passPriorityAroundTable(t, g)
		life := me.Life
		pr7Hit(t, g, theirs, me.ID, 3)
		if me.Life != life {
			t.Fatalf("life %d, want %d", me.Life, life)
		}
	})
	t.Run("Winds of Qal Sisma", func(t *testing.T) {
		for _, ferocious := range []bool{false, true} {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			power := 3
			if ferocious {
				power = 4
			}
			att := sfPush(g, me.ID, "Attacker", "Creature — Test", power, []string{"G"})
			blk := sfPush(g, opp.ID, "Blocker", "Creature — Test", 2, []string{"R"})
			pr7bAttack(t, g, opp.ID, att)
			pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blk: att})
			sfCast(t, g, me, sfInstant("Winds of Qal Sisma", sfWindsOfQalSisma), game.CastSpellParams{})
			sfCombat(t, g)
			wantBlk := 0
			if ferocious {
				wantBlk = 4
			}
			if pr6Marked(g, blk) != wantBlk || pr6Marked(g, att) != 0 {
				t.Fatalf("ferocious %v: blocker %d (want %d), attacker %d (want 0)", ferocious, pr6Marked(g, blk), wantBlk, pr6Marked(g, att))
			}
		}
	})
	t.Run("Comeuppance", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := sfPush(g, me.ID, "A", "Creature — Test", 3, []string{"R"})
		b := sfPush(g, me.ID, "B", "Creature — Test", 2, []string{"R"})
		rock := sfPush(g, me.ID, "Rock", "Artifact", 0, nil)
		pr7bAttack(t, g, opp.ID, a, b)
		sfCast(t, g, opp, sfInstant("Comeuppance", sfComeuppance), game.CastSpellParams{})
		lost := sfCombat(t, g)
		if lost[opp.ID] != 0 || pr6Marked(g, a) != 3 || pr6Marked(g, b) != 2 {
			t.Fatalf("defender lost %d (want 0); A %d (want 3), B %d (want 2): each creature is dealt its own damage",
				lost[opp.ID], pr6Marked(g, a), pr6Marked(g, b))
		}
		life := me.Life
		pr7Hit(t, g, rock, opp.ID, 2)
		if me.Life != life-2 {
			t.Fatalf("source's controller life %d, want %d: a noncreature source's damage goes to its controller", me.Life, life-2)
		}
	})
	t.Run("Judgment of Alexander", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		apaPush(g, opp.ID, opp.ID, game.Card{Name: "Commander", TypeLine: "Legendary Creature — Test", Power: 4, Toughness: 4, IsCommander: true})
		a := sfPush(g, me.ID, "A", "Creature — Test", 3, []string{"R"})
		b := sfPush(g, me.ID, "B", "Creature — Test", 2, []string{"R"})
		pr7bAttack(t, g, opp.ID, a, b)
		sfCast(t, g, opp, sfInstant("Judgment of Alexander", sfJudgment), game.CastSpellParams{})
		lost := sfCombat(t, g)
		if lost[opp.ID] != 0 || pr6Marked(g, a) != 4 || pr6Marked(g, b) != 4 {
			t.Fatalf("defender lost %d (want 0); A %d, B %d (want the commander's 4 each)", lost[opp.ID], pr6Marked(g, a), pr6Marked(g, b))
		}
	})
}

// Objects pinned as the shield resolves: Terrifying Presence's target,
// Haze Frog itself; Lithomancer's Focus's colourless sources and
// Frontline Strategist's non-Soldiers.
func TestShieldFilterCardsPinnedAndProtected(t *testing.T) {
	t.Run("Terrifying Presence", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		chosen := sfPush(g, me.ID, "Chosen", "Creature — Test", 3, []string{"G"})
		other := sfPush(g, me.ID, "Other", "Creature — Test", 2, []string{"G"})
		pr7bAttack(t, g, opp.ID, chosen, other)
		sfCast(t, g, me, game.Card{Name: "Terrifying Presence", TypeLine: "Instant", OracleID: sfTerrifyingPresence},
			game.CastSpellParams{Targets: pr7bTargets(chosen)})
		if lost := sfCombat(t, g)[opp.ID]; lost != 3 {
			t.Fatalf("defender lost %d, want 3: only the target deals damage", lost)
		}
	})
	t.Run("Haze Frog", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := sfPush(g, me.ID, "A", "Creature — Test", 3, []string{"R"})
		b := sfPush(g, me.ID, "B", "Creature — Test", 2, []string{"R"})
		sfAttackHold(t, g, opp.ID, a, b)
		frog := sfCast(t, g, opp, game.Card{Name: "Haze Frog", TypeLine: "Creature — Frog", OracleID: sfHazeFrog,
			Power: 2, Toughness: 1}, game.CastSpellParams{})
		pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{frog: a})
		lost := sfCombat(t, g)
		if lost[opp.ID] != 0 || pr6Marked(g, a) != 2 || pr6Marked(g, frog) != 0 {
			t.Fatalf("defender lost %d (want 0); A %d (want the Frog's 2), Frog %d (want 0)", lost[opp.ID], pr6Marked(g, a), pr6Marked(g, frog))
		}
	})
	t.Run("Lithomancer's Focus", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := sfPush(g, me.ID, "Mine", "Creature — Test", 1, []string{"W"})
		bystander := sfPush(g, me.ID, "Bystander", "Creature — Test", 1, []string{"W"})
		golem := sfPush(g, opp.ID, "Golem", "Artifact Creature — Golem", 2, nil)
		red := sfPush(g, opp.ID, "Red", "Creature — Test", 2, []string{"R"})
		castCatalogSpell(t, g, "Lithomancer's Focus", "Instant", sfLithomancersFocus, pr7bTargets(mine))
		passPriorityAroundTable(t, g)
		if p := findBattlefieldCardForTest(g, mine).CurrentPower(); p != 3 {
			t.Fatalf("power %d, want 3", p)
		}
		pr6Damage(t, g, golem, mine, 2)
		pr6Damage(t, g, red, mine, 1)
		pr6Damage(t, g, golem, bystander, 1)
		if pr6Marked(g, mine) != 1 || pr6Marked(g, bystander) != 1 {
			t.Fatalf("target %d (want the red source's 1), bystander %d (want 1)", pr6Marked(g, mine), pr6Marked(g, bystander))
		}
	})
	t.Run("Frontline Strategist", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		id := castWithAltCost(t, g, "Frontline Strategist", "Creature — Human Soldier", sfFrontlineStrategist, "morph")
		passPriorityAroundTable(t, g)
		soldier := sfPush(g, me.ID, "Soldier", "Creature — Human Soldier", 3, []string{"W"})
		bear := sfPush(g, me.ID, "Bear", "Creature — Bear", 2, []string{"G"})
		g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens("W")...) })
		if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
			t.Fatalf("turn face up: %v", err)
		}
		passPriorityAroundTable(t, g)
		pr7bAttack(t, g, opp.ID, soldier, bear)
		if lost := sfCombat(t, g)[opp.ID]; lost != 3 {
			t.Fatalf("defender lost %d, want 3: only the Soldier deals damage", lost)
		}
	})
}
