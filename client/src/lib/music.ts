// Ambient music playback — a single long-form track that loops
// forever as the app shell's background bed.
//
// Deliberately a separate module from sounds.ts because the SFX
// model (clone-on-play pools, per-event rate limits, random variant
// picker) doesn't translate: music is one element, always the same
// URL, and only ever has one playing instance. Sharing the pool
// machinery would be more code, not less.
//
// Autoplay policy: same constraint as SFX — HTMLAudioElement.play()
// rejects until the page sees a user gesture. armMusicOnFirstGesture()
// installs a one-shot pointerdown/keydown listener that starts the
// track. Safe to call many times.
//
// Volume/mute: routed through settings.ts the same way SFX volume
// is. master × music (both 0..100 in the store) folds into a 0..1
// multiplier here; the audio.muted toggle pauses/resumes the track
// (vs. the SFX path which just no-ops play()). We pause rather than
// zero the volume so a muted user isn't silently decoding audio
// they'll never hear.
//
// Dependency-free + SSR-safe (window-guarded) so it can be imported
// from any surface.

const AMBIENT_TRACK_URL = "/sounds/ambient_bronze_thunder.mp3";

let audio: HTMLAudioElement | null = null;
let armed = false;
// musicMultiplier is master × music volume, both 0..100 in the
// settings store, normalised to 0..1 here. The settings bridge
// (settings.ts) pushes updates via setMusicVolumeMultiplier.
let musicMultiplier = 0;
let muted = false;
// duck is 1, or about -6 dB while an interface cue plays (ADR 0209 §2).
// sounds.ts reports it through onDuck, wired in settings.ts.
let duck = 1;

function hasWindow(): boolean {
  return typeof window !== "undefined";
}

// The Audio element (and so the 6 MB download) is created lazily, the
// first time music is armed and actually wanted: never while muted or
// at volume 0.
function ensureAudio(): HTMLAudioElement {
  if (!audio) {
    audio = new Audio(AMBIENT_TRACK_URL);
    audio.loop = true;
    audio.preload = "auto";
  }
  return audio;
}

function applyVolume(): void {
  if (!armed) return;
  if (muted || musicMultiplier === 0) {
    // Pause when muted so we're not decoding a silent stream in the
    // background tab. Unpause picks back up at the last position,
    // which is the right UX for an ambient bed.
    if (audio && !audio.paused) audio.pause();
    return;
  }
  const a = ensureAudio();
  a.volume = musicMultiplier * duck;
  if (a.paused) void a.play().catch(() => {});
}

export function setMusicVolumeMultiplier(v: number): void {
  musicMultiplier = Math.min(1, Math.max(0, v));
  applyVolume();
}

// setMusicDuck scales the music bus while an interface cue plays: 1 for
// no change, 0.5 for about 6 dB down. Applies to a track that is playing
// and to one that starts later.
export function setMusicDuck(factor: number): void {
  duck = Math.min(1, Math.max(0, factor));
  if (audio && armed && !muted && musicMultiplier > 0) audio.volume = musicMultiplier * duck;
}

export function setMusicMuted(m: boolean): void {
  muted = m;
  applyVolume();
}

// armMusicOnFirstGesture attaches a one-shot pointerdown + keydown
// listener to window. On first of either, it starts playback if music
// is wanted (not muted, volume above 0); otherwise the Audio element
// is not created until it is. Subsequent calls no-op so the App
// shell can re-run it from a reactive effect without side effects.
export function armMusicOnFirstGesture(): void {
  if (armed || !hasWindow()) return;
  armed = true;
  const start = () => {
    window.removeEventListener("pointerdown", start);
    window.removeEventListener("keydown", start);
    applyVolume();
  };
  window.addEventListener("pointerdown", start, { once: true });
  window.addEventListener("keydown", start, { once: true });
}
