package testsso

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// spectral_searchlight_test.go — #6101: "{T}: Choose a player. That
// player adds one mana of any color they choose." A mana ability (CR
// 605.1a) whose activator names the player and whose chosen player
// names the colour and gets the mana.

const spectralSearchlightOracle = "511b1174-8ef7-4d17-93de-01b9b81bda6a"

func pushSearchlight(g *game.Game, controller uuid.UUID) uuid.UUID {
	return effectstest.PushPermanentForTest(g, controller, "Spectral Searchlight", spectralSearchlightOracle, "Artifact")
}

func searchlightTapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c := effectstest.FindBattlefieldCardForTest(g, id)
	if c == nil {
		t.Fatalf("Spectral Searchlight is not on the battlefield")
	}
	return c.Tapped
}

func pickForPlayer(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	return effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceManaForPlayer, chooser)
}

func TestSpectralSearchlightRowChoosesAPlayer(t *testing.T) {
	rows := game.ManaAbilitiesForCard(game.Card{OracleID: spectralSearchlightOracle, Name: "Spectral Searchlight", TypeLine: "Artifact"})
	if len(rows) != 1 {
		t.Fatalf("mana rows = %d, want 1", len(rows))
	}
	ab := rows[0]
	if !ab.ChosenPlayerAdds || !ab.TapCost || ab.Produced != "{W|U|B|R|G}" || ab.AnyPlayer {
		t.Errorf("row = %+v, want \"{T}: choose a player, that player adds one mana of any color\"", ab)
	}
}

// Choosing yourself is Birds of Paradise: a mana_pick for the activator,
// or the colour named up front.
func TestSpectralSearchlightChoosingYourself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	light := pushSearchlight(g, me.ID)

	if err := g.ActivateManaAbility(me.ID, light, 0, game.ManaAbilityParams{ChosenPlayer: me.ID}); err != nil {
		t.Fatalf("activate choosing myself: %v", err)
	}
	if !searchlightTapped(t, g, light) {
		t.Fatal("Searchlight not tapped")
	}
	if pickForPlayer(g, me.ID) != nil {
		t.Fatal("choosing yourself asked a mana_pick_for_player")
	}
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceMana, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("no five-colour mana_pick for the activator: %+v", g.PendingChoices)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "B"); err != nil {
		t.Fatalf("answer: %v", err)
	}
	if me.ManaPool.CountColor("B") != 1 || me.ManaPool.Total() != 1 || opp.ManaPool.Total() != 0 {
		t.Errorf("pools: me %v opp %v, want one {B} in mine", me.ManaPool, opp.ManaPool)
	}

	// Untap it and name the colour up front, with no player named: the
	// activator, with no prompt at all.
	effectstest.FindBattlefieldCardForTest(g, light).Tapped = false
	if err := g.ActivateManaAbility(me.ID, light, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("activate with an up-front colour: %v", err)
	}
	if me.ManaPool.CountColor("G") != 1 || len(g.PendingChoices) != 0 {
		t.Errorf("pool %v, choices %+v: want {G} added at once", me.ManaPool, g.PendingChoices)
	}
}

// Choosing an opponent pays the cost, asks the opponent for the colour,
// and the mana goes to the opponent's pool.
func TestSpectralSearchlightChoosingAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	light := pushSearchlight(g, me.ID)

	if err := g.ActivateManaAbility(me.ID, light, 0, game.ManaAbilityParams{ChosenPlayer: opp.ID}); err != nil {
		t.Fatalf("activate choosing an opponent: %v", err)
	}
	if !searchlightTapped(t, g, light) {
		t.Fatal("the cost was not paid")
	}
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceMana, me.ID) != nil {
		t.Fatal("the activator was asked the colour")
	}
	pick := pickForPlayer(g, opp.ID)
	if pick == nil {
		t.Fatalf("no mana_pick_for_player for the chosen opponent: %+v", g.PendingChoices)
	}
	if pick.FromPlayer != me.ID || pick.Source != light || len(pick.ColorOptions) != 5 {
		t.Errorf("pick = %+v, want from the activator, sourced from the Searchlight, five colours", pick)
	}
	if !game.ChoiceBlocksTable(pick.Kind) {
		t.Error("the pick does not hold the table: the ability has not finished resolving")
	}
	if me.ManaPool.Total() != 0 || opp.ManaPool.Total() != 0 {
		t.Fatalf("mana added before the colour was chosen: me %v opp %v", me.ManaPool, opp.ManaPool)
	}

	// Only the chosen player answers, and only with a colour on offer.
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); !errors.Is(err, game.ErrNotTheChooser) {
		t.Errorf("the activator answered: err = %v, want ErrNotTheChooser", err)
	}
	if err := g.ResolveManaChoice(pick.ID, opp.ID, "C"); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("colourless answered: err = %v, want ErrInvalidParam", err)
	}

	// The opponent's colour choice: their pool, the Searchlight's mana.
	if err := g.ResolveManaChoice(pick.ID, opp.ID, "U"); err != nil {
		t.Fatalf("the opponent answers: %v", err)
	}
	if opp.ManaPool.CountColor("U") != 1 || opp.ManaPool.Total() != 1 {
		t.Errorf("opponent's pool = %v, want one {U}", opp.ManaPool)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("activator's pool = %v, want empty", me.ManaPool)
	}
	if tok := opp.ManaPool.Units(); len(tok) != 1 || tok[0].Source != light {
		t.Errorf("tokens = %+v, want one sourced from the Searchlight", tok)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("choices left open: %+v", g.PendingChoices)
	}
}

// Refusals, before anything is paid: a player not in the game, a colour
// named for someone else, and chosen_player on an ability that chooses
// no player.
func TestSpectralSearchlightRefusalsCostNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	light := pushSearchlight(g, me.ID)

	if err := g.ActivateManaAbility(me.ID, light, 0, game.ManaAbilityParams{ChosenPlayer: uuid.New()}); !errors.Is(err, game.ErrPlayerNotFound) {
		t.Errorf("a stranger: err = %v, want ErrPlayerNotFound", err)
	}
	if err := g.ActivateManaAbility(me.ID, light, 0, game.ManaAbilityParams{ChosenPlayer: opp.ID, Colors: []string{"R"}}); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("a colour named for the opponent: err = %v, want ErrInvalidParam", err)
	}
	if err := g.Concede(third.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, light, 0, game.ManaAbilityParams{ChosenPlayer: third.ID}); !errors.Is(err, game.ErrPlayerNotFound) {
		t.Errorf("a player who has left: err = %v, want ErrPlayerNotFound", err)
	}
	if searchlightTapped(t, g, light) || len(g.PendingChoices) != 0 {
		t.Fatalf("a refused activation paid or asked: tapped=%v choices=%+v", searchlightTapped(t, g, light), g.PendingChoices)
	}

	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	if err := g.ActivateManaAbility(me.ID, forest, 0, game.ManaAbilityParams{ChosenPlayer: opp.ID}); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("chosen_player on a Forest: err = %v, want ErrInvalidParam", err)
	}
}

// A chosen player who leaves the game with the pick open (a disconnected
// player who concedes, or is removed) takes the question with them: the
// mana would enter a pool that has left (CR 800.4a), and the table moves
// on. The cost stays paid.
func TestSpectralSearchlightChosenPlayerLeavesWithThePickOpen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	light := pushSearchlight(g, me.ID)
	if err := g.ActivateManaAbility(me.ID, light, 0, game.ManaAbilityParams{ChosenPlayer: opp.ID}); err != nil {
		t.Fatal(err)
	}
	if pickForPlayer(g, opp.ID) == nil {
		t.Fatal("no pick")
	}
	if err := g.Concede(opp.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("the departed player's pick is still open: %+v", g.PendingChoices)
	}
	if me.ManaPool.Total() != 0 || !searchlightTapped(t, g, light) {
		t.Errorf("pool %v tapped %v: want no mana and the cost paid", me.ManaPool, searchlightTapped(t, g, light))
	}
}

// A disconnected chooser answers when they are back: the open pick is
// part of the saved game, and a restored table asks it again.
func TestSpectralSearchlightPickSurvivesARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	light := pushSearchlight(g, me.ID)
	if err := g.ActivateManaAbility(me.ID, light, 0, game.ManaAbilityParams{ChosenPlayer: opp.ID}); err != nil {
		t.Fatal(err)
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	me, opp = g.Seats[0], g.Seats[1]
	pick := pickForPlayer(g, opp.ID)
	if pick == nil || pick.FromPlayer != me.ID {
		t.Fatalf("restored pick = %+v, want one for the opponent from the activator", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, opp.ID, "W"); err != nil {
		t.Fatal(err)
	}
	if opp.ManaPool.CountColor("W") != 1 || me.ManaPool.Total() != 0 {
		t.Errorf("pools: me %v opp %v, want one {W} in the opponent's", me.ManaPool, opp.ManaPool)
	}
}

// A bot (or any agent seat) chosen is offered the colours as moves, and
// nothing else while the pick is open; the activator is offered no
// answer.
func TestSpectralSearchlightChosenSeatIsOfferedTheColours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	light := pushSearchlight(g, me.ID)
	if err := g.ActivateManaAbility(me.ID, light, 0, game.ManaAbilityParams{ChosenPlayer: opp.ID}); err != nil {
		t.Fatal(err)
	}
	pick := pickForPlayer(g, opp.ID)
	colors := map[string]bool{}
	for _, m := range legal.EnumerateFor(g, opp.ID) {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("the chosen seat is offered %s %q while its pick is open", m.Type, m.Label)
			continue
		}
		var p struct {
			ChoiceID string `json:"choice_id"`
			Color    string `json:"color"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if p.ChoiceID != pick.ID.String() {
			t.Errorf("answer for choice %s, want %s", p.ChoiceID, pick.ID)
		}
		colors[p.Color] = true
	}
	if len(colors) != 5 {
		t.Errorf("colours offered = %v, want all five", colors)
	}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == legal.TypeResolveChoice {
			t.Errorf("the activator is offered an answer to the opponent's pick: %q", m.Label)
		}
	}
}

// Use mid-payment (CR 605.3a): a Rhystic Study tax is open for the
// activator. Choosing yourself floats the mana that pays it; choosing
// another player leaves the tax unpaid and that player with the mana.
func TestSpectralSearchlightMidPayment(t *testing.T) {
	t.Run("choosing yourself pays the tax", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		caster, owner := g.Seats[0], g.Seats[1]
		light := pushSearchlight(g, caster.ID)
		effectstest.PushPermanentForTest(g, owner.ID, "Rhystic Study", effectstest.RhysticStudyOracle, "Enchantment")
		effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasPayUnlessFor(g, caster.ID) {
			t.Fatalf("no Rhystic tax for the caster: %+v", g.PendingChoices)
		}
		if err := g.ActivateManaAbility(caster.ID, light, 0, game.ManaAbilityParams{ChosenPlayer: caster.ID, Colors: []string{"U"}}); err != nil {
			t.Fatalf("activate mid-payment: %v", err)
		}
		hand := owner.Hand.Size()
		effectstest.AnswerPayUnless(t, g, caster.ID, true)
		if caster.ManaPool.Total() != 0 {
			t.Errorf("the tax did not spend the floated mana: %v", caster.ManaPool)
		}
		if effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm) != nil || owner.Hand.Size() != hand {
			t.Error("the paid tax still offered the draw")
		}
	})
	t.Run("choosing another player gives them the mana", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		caster, owner := g.Seats[0], g.Seats[1]
		light := pushSearchlight(g, caster.ID)
		effectstest.PushPermanentForTest(g, owner.ID, "Rhystic Study", effectstest.RhysticStudyOracle, "Enchantment")
		effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
		effectstest.PassPriorityAroundTable(t, g)
		if err := g.ActivateManaAbility(caster.ID, light, 0, game.ManaAbilityParams{ChosenPlayer: owner.ID}); err != nil {
			t.Fatalf("activate mid-payment: %v", err)
		}
		pick := pickForPlayer(g, owner.ID)
		if pick == nil || !effectstest.HasPayUnlessFor(g, caster.ID) {
			t.Fatalf("want the owner's pick and the caster's tax both open: %+v", g.PendingChoices)
		}
		if err := g.ResolveManaChoice(pick.ID, owner.ID, "R"); err != nil {
			t.Fatal(err)
		}
		if owner.ManaPool.CountColor("R") != 1 || caster.ManaPool.Total() != 0 {
			t.Errorf("pools: caster %v owner %v, want one {R} in the owner's", caster.ManaPool, owner.ManaPool)
		}
		if !effectstest.HasPayUnlessFor(g, caster.ID) {
			t.Error("the caster's tax closed")
		}
	})
}

// The auto-tapper never plans it: a strict auto-tapped cast with only
// the Searchlight to pay is refused and leaves it untapped; with a
// Mountain beside it the Mountain pays.
func TestSpectralSearchlightIsNeverAutoTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	light := pushSearchlight(g, me.ID)
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	bolt := func() uuid.UUID {
		id := uuid.New()
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}",
			OracleID: effectstest.LightningBoltOracle, Owner: me.ID, Controller: me.ID,
		})
		return id
	}
	opp := g.Seats[1].ID
	params := game.CastSpellParams{Strict: true, AutoTap: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp}}}
	if err := g.CastSpell(me.ID, bolt(), params); err == nil {
		t.Fatal("a cast only the Searchlight could pay was accepted")
	}
	if searchlightTapped(t, g, light) {
		t.Fatal("the auto-tapper tapped the Searchlight")
	}
	mountain := effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	if err := g.CastSpell(me.ID, bolt(), params); err != nil {
		t.Fatalf("cast with a Mountain: %v", err)
	}
	if searchlightTapped(t, g, light) || !effectstest.FindBattlefieldCardForTest(g, mountain).Tapped {
		t.Error("want the Mountain tapped and the Searchlight not")
	}
}
