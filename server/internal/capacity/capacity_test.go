package capacity

import (
	"strings"
	"testing"
	"time"
)

func env(kv map[string]string) func(string) string {
	return func(k string) string { return kv[k] }
}

func TestDefaultsAreTheOwnersDecision(t *testing.T) {
	l, err := FromEnv(env(nil))
	if err != nil {
		t.Fatal(err)
	}
	if l != Defaults() {
		t.Errorf("an empty environment must give the defaults:\n got %+v\nwant %+v", l, Defaults())
	}
	// 30 games of four humans (120 players) is the owner's soft cap
	// (#2977, 2026-10-11) and fits exactly.
	if got, want := l.TableCost(4, 0)*30, l.MaxLoad; got != want {
		t.Errorf("30 four-human tables cost %d, cap is %d: they should fit exactly", got, want)
	}
	// A bot seat weighs more than a human seat, and the bot-table
	// ceiling stays well under the ~31 four-bot tables prod fell over at.
	if l.BotSeatWeight <= l.HumanSeatWeight {
		t.Errorf("bot weight %d must exceed human weight %d", l.BotSeatWeight, l.HumanSeatWeight)
	}
	if per := l.TableCost(0, 4); l.MaxLoad/per >= 31 || l.MaxLoad/per < 15 {
		t.Errorf("the cap holds %d four-bot tables; want between 15 and 30", l.MaxLoad/per)
	}
}

func TestFromEnvReadsEveryVariable(t *testing.T) {
	l, err := FromEnv(env(map[string]string{
		EnvMaxLoad:      "90",
		EnvWeightTable:  "3",
		EnvWeightHuman:  "2",
		EnvWeightBot:    "5",
		EnvMaxConns:     "250",
		EnvAccountBurst: "4",
		EnvAccountEvery: "45s",
		EnvIPBurst:      "8",
		EnvIPEvery:      "1m",
		EnvIdleLobby:    "12h",
		EnvIdleGame:     "72h",
		EnvIdleBots:     "30m",
	}))
	if err != nil {
		t.Fatal(err)
	}
	want := Limits{MaxLoad: 90, TableWeight: 3, HumanSeatWeight: 2, BotSeatWeight: 5, MaxConns: 250,
		AccountBurst: 4, AccountEvery: 45 * time.Second, IPBurst: 8, IPEvery: time.Minute,
		IdleLobbyAfter: 12 * time.Hour, IdleGameAfter: 72 * time.Hour, IdleBotsAfter: 30 * time.Minute}
	if l != want {
		t.Errorf("got %+v\nwant %+v", l, want)
	}
	if got := l.TableCost(1, 3); got != 3+2+15 {
		t.Errorf("TableCost with custom weights = %d", got)
	}
}

func TestZeroSwitchesACapOff(t *testing.T) {
	l, err := FromEnv(env(map[string]string{EnvMaxLoad: "0", EnvMaxConns: "0"}))
	if err != nil {
		t.Fatal(err)
	}
	if l.LoadCapped() || !l.Fits(1<<30, 1<<30) {
		t.Error("MaxLoad 0 must be no cap")
	}
	if l.MaxConns != 0 {
		t.Error("MaxConns 0 must be no cap")
	}
}

func TestAMalformedValueFailsNamingTheVariable(t *testing.T) {
	for k, v := range map[string]string{
		EnvMaxLoad:      "lots",
		EnvMaxConns:     "-1",
		EnvWeightBot:    "2.5",
		EnvAccountBurst: "0", // a burst of zero would refuse every create forever
		EnvIPBurst:      "0",
		EnvAccountEvery: "soon",
		EnvIPEvery:      "0s",
		EnvIdleLobby:    "a day",
		EnvIdleGame:     "-48h",
		EnvIdleBots:     "60", // a bare number has no unit
	} {
		_, err := FromEnv(env(map[string]string{k: v}))
		if err == nil {
			t.Errorf("%s=%q was accepted", k, v)
			continue
		}
		if !strings.Contains(err.Error(), k) {
			t.Errorf("%s=%q: error %q does not name the variable", k, v, err)
		}
	}
}

func TestFits(t *testing.T) {
	l := Limits{MaxLoad: 10}
	if !l.Fits(6, 4) || l.Fits(6, 5) {
		t.Error("Fits is used+cost <= cap")
	}
}

func TestWithCreateDefaultsFillsOnlyTheZeros(t *testing.T) {
	got := Limits{AccountBurst: 9}.WithCreateDefaults()
	d := Defaults()
	if got.AccountBurst != 9 || got.AccountEvery != d.AccountEvery || got.IPBurst != d.IPBurst || got.IPEvery != d.IPEvery {
		t.Errorf("%+v", got)
	}
}

// The sentence and the code are the wire contract; the client (Go
// constants here, TypeScript in lib/serverFull.ts) matches the code.
func TestTheRefusalIsAStableCodeAndAFriendlySentence(t *testing.T) {
	if CodeServerFull != "server_full" {
		t.Errorf("the code the client matches changed: %q", CodeServerFull)
	}
	if ErrServerFull.Error() != Message {
		t.Error("ErrServerFull must read as the friendly sentence")
	}
	for _, bad := range []string{"error", "failed", "500", "503", "overload", "exception"} {
		if strings.Contains(strings.ToLower(Message), bad) {
			t.Errorf("the message %q must not read like a raw error (%q)", Message, bad)
		}
	}
}

// ADR 0211 §7: the owner's thresholds are the defaults, and "0" (or a
// zero duration) switches each cleanup off on its own.
func TestIdleCleanupDefaultsAndOff(t *testing.T) {
	d := Defaults()
	if d.IdleLobbyAfter != 24*time.Hour || d.IdleGameAfter != 48*time.Hour || d.IdleBotsAfter != time.Hour {
		t.Errorf("idle defaults %v / %v / %v, want 24h / 48h / 1h", d.IdleLobbyAfter, d.IdleGameAfter, d.IdleBotsAfter)
	}
	l, err := FromEnv(env(map[string]string{EnvIdleLobby: "0", EnvIdleGame: "0s", EnvIdleBots: ""}))
	if err != nil {
		t.Fatal(err)
	}
	if l.IdleLobbyAfter != 0 || l.IdleGameAfter != 0 {
		t.Errorf("0 must switch the cleanup off: lobby %v, game %v", l.IdleLobbyAfter, l.IdleGameAfter)
	}
	if l.IdleBotsAfter != time.Hour {
		t.Errorf("an empty value is the default: bots %v", l.IdleBotsAfter)
	}
	if (Limits{}).IdleLobbyAfter != 0 {
		t.Error("the zero Limits must have every cleanup off")
	}
}
