package api

import "net/http"

// Public legal + support pages (English + German), linked from the app and used
// for the App Store privacy-policy / EULA / support URLs. Solid DSGVO template —
// have it reviewed and add a full postal address (Impressum) before release.

func (a *API) handleTerms(w http.ResponseWriter, r *http.Request)     { writeHTML(w, legalDoc("en", "Terms of Use", navTerms, termsEN)) }
func (a *API) handleTermsDE(w http.ResponseWriter, r *http.Request)   { writeHTML(w, legalDoc("de", "Nutzungsbedingungen", navTerms, termsDE)) }
func (a *API) handlePrivacy(w http.ResponseWriter, r *http.Request)   { writeHTML(w, legalDoc("en", "Privacy Policy", navPrivacy, privacyEN)) }
func (a *API) handlePrivacyDE(w http.ResponseWriter, r *http.Request) { writeHTML(w, legalDoc("de", "Datenschutzerklärung", navPrivacy, privacyDE)) }
func (a *API) handleSupport(w http.ResponseWriter, r *http.Request)   { writeHTML(w, legalDoc("en", "Support", navSupport, supportEN)) }
func (a *API) handleSupportDE(w http.ResponseWriter, r *http.Request) { writeHTML(w, legalDoc("de", "Support", navSupport, supportDE)) }
func (a *API) handleImpressum(w http.ResponseWriter, r *http.Request) { writeHTML(w, legalDoc("de", "Impressum", navImpressum, impressumDE)) }

func writeHTML(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(body))
}

func legalDoc(lang, title, nav, inner string) string {
	return `<!doctype html><html lang="` + lang + `"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Inhale With Me — ` + title + `</title>
<style>
:root{--ink:#1c1b19;--muted:#6b6459;--accent:#c2560f;--bg:#f4f1ea;--card:#fffdf8;--line:#e4ddd0}
*{box-sizing:border-box}
body{font:16px/1.65 -apple-system,BlinkMacSystemFont,"Segoe UI",system-ui,sans-serif;color:var(--ink);background:var(--bg);margin:0}
.wrap{max-width:720px;margin:0 auto;padding:36px 20px 90px}
.top{display:flex;justify-content:space-between;align-items:center;gap:12px;flex-wrap:wrap;
  border-bottom:1px solid var(--line);padding-bottom:16px;margin-bottom:28px}
.brand{display:flex;align-items:center;gap:10px;font-weight:700;letter-spacing:-.01em;font-size:15px}
.mark{width:22px;height:22px;border-radius:6px;background:linear-gradient(150deg,#2a2018,#12100c);position:relative;flex:none}
.mark::before{content:"";position:absolute;width:3px;height:11px;border-radius:2px;background:linear-gradient(#e5a24c,#f4ede0);transform:rotate(38deg);left:9px;top:5px}
.mark::after{content:"";position:absolute;width:3px;height:3px;border-radius:50%;background:var(--accent);box-shadow:0 0 6px 1px var(--accent);left:6px;top:13px}
.nav{display:flex;gap:6px}
.nav a{font-size:13px;font-weight:600;text-decoration:none;color:var(--muted);border:1px solid var(--line);
  background:var(--card);padding:5px 11px;border-radius:999px}
.nav a.on{color:#fff;background:var(--accent);border-color:var(--accent)}
.card{background:var(--card);border:1px solid var(--line);border-radius:16px;padding:26px 30px}
h1{font-size:30px;line-height:1.15;margin:0 0 6px}
h2{font-size:17px;margin:30px 0 6px;color:var(--accent)}
.date{color:var(--muted);font-size:14px;margin:0 0 4px}
ul{padding-left:1.15em}li{margin:4px 0}
a{color:var(--accent)}
.muted{color:var(--muted);font-size:13px;margin-top:26px;border-top:1px solid var(--line);padding-top:14px}
</style></head><body><div class="wrap">
<div class="top">
  <span class="brand"><span class="mark"></span>Inhale With Me</span>
  <nav class="nav">` + nav + `</nav>
</div>
<div class="card">` + inner + `</div>
</div></body></html>`
}

const navPrivacy = `<a href="/privacy">EN</a><a href="/privacy/de">DE</a>`
const navTerms = `<a href="/terms">EN</a><a href="/terms/de">DE</a>`
const navSupport = `<a href="/support">EN</a><a href="/support/de">DE</a>`
const navImpressum = `<a href="/privacy/de">Datenschutz</a><a href="/support/de">Support</a>`

const impressumDE = `
<h1>Impressum</h1>
<p>Angaben gemäß § 5 DDG</p>
<p>Noah Feldt<br>Niedertorstr. 32<br>32312 Lübbecke<br>Deutschland</p>

<h2>Kontakt</h2>
<p>E-Mail: <a href="mailto:noah@feldt.systems">noah@feldt.systems</a></p>

<h2>Verantwortlich für den Inhalt (§ 18 Abs. 2 MStV)</h2>
<p>Noah Feldt, Anschrift wie oben.</p>`

// ---------------------------------------------------------------- English

const privacyEN = `
<h1>Privacy Policy</h1>
<p class="date">Last updated: 2026</p>

<h2>1. Controller</h2>
<p>Responsible for data processing in this app:<br>Noah Feldt · Niedertorstr. 32, 32312 Lübbecke, Germany · noah@feldt.systems</p>

<h2>2. Overview</h2>
<p>Inhale With Me (the “App”) lets you log smoke sessions and share them with friends. This policy explains what personal data we process and why.</p>

<h2>3. What we process</h2>
<ul>
<li><strong>Account:</strong> email, username, optional display name, bio, avatar URL.</li>
<li><strong>Activity you log:</strong> session type, quantity, time, optional note/mood/location, and visibility.</li>
<li><strong>Social:</strong> friends, friend requests, reactions.</li>
<li><strong>Device token:</strong> an Apple push token, only if you enable notifications.</li>
<li><strong>Location (optional):</strong> only when you actively share it, your coordinates are sent to the chosen friend and stored briefly (about 1 hour), then discarded. No background tracking.</li>
<li><strong>Technical:</strong> IP address and server logs when the app calls our API.</li>
</ul>

<h2>4. Purposes and legal bases (Art. 6 GDPR)</h2>
<ul>
<li>Providing the app and your account: Art. 6(1)(b) (contract).</li>
<li>Push notifications: Art. 6(1)(a) (consent); revocable in iOS Settings.</li>
<li>Location sharing: Art. 6(1)(a) (consent), only when you choose to share.</li>
<li>Operation, security, abuse prevention, logs: Art. 6(1)(f) (legitimate interests).</li>
</ul>

<h2>5. Recipients</h2>
<p>Hosting: mittwald (servers in Germany/EU). Push delivery: Apple Push Notification service (Apple Inc.). We do not sell your data or share it for advertising.</p>

<h2>6. Visibility of social content</h2>
<p>Your sessions are visible only to you and, per the visibility you choose (public / friends / private), to your friends.</p>

<h2>7. Retention &amp; deletion</h2>
<p>We keep your data until you delete it. You can delete your account and all associated data at any time in the app under Profile → Delete Account.</p>

<h2>8. Your rights</h2>
<p>You have the right to access, rectification, erasure, restriction, data portability and objection, and to withdraw consent at any time. Contact: noah@feldt.systems. You also have the right to lodge a complaint with a data protection authority.</p>

<h2>9. Children</h2>
<p>The App is for adults (17+) and is not directed to children.</p>

<h2>10. Changes</h2>
<p>We may update this policy; the current version always applies.</p>

<p class="muted">Please have this reviewed by a professional before public release.</p>`

const termsEN = `
<h1>Terms of Use (EULA)</h1>
<p class="date">Last updated: 2026</p>
<p>By creating an account or using Inhale With Me (the “App”), you agree to these Terms. If you do not agree, do not use the App.</p>

<h2>1. Eligibility</h2>
<p>You must be at least 18 years old (or the age of majority where you live). The App is a personal tracking and social tool; it does not sell, promote, or encourage the consumption of any product.</p>

<h2>2. Your account</h2>
<p>You are responsible for your credentials and for the content you log and share.</p>

<h2>3. Acceptable use</h2>
<p>Do not misuse the App, harass others, or post unlawful or objectionable content. There is zero tolerance for objectionable content or abusive behavior. You can report content and block users in the App; reports are reviewed and offending content or accounts may be removed.</p>

<h2>4. User content</h2>
<p>You keep ownership of what you create and grant us a limited license to store and display it to you and, per your visibility settings, to your friends, solely to operate the App.</p>

<h2>5. Account deletion</h2>
<p>You can delete your account at any time under Profile → Delete Account, which permanently removes your data.</p>

<h2>6. Disclaimer</h2>
<p>The App is provided “as is,” without warranties, and is not medical advice. To the extent permitted by law, we are not liable for damages arising from your use of the App.</p>

<h2>7. Changes &amp; contact</h2>
<p>We may update these Terms; continued use constitutes acceptance. Contact: noah@feldt.systems</p>`

const supportEN = `
<h1>Support</h1>
<p>Need help with Inhale With Me? We're happy to help.</p>

<h2>Contact</h2>
<p>Email <a href="mailto:noah@feldt.systems">noah@feldt.systems</a>. We aim to reply within a few days.</p>

<h2>Common questions</h2>
<ul>
<li><strong>Delete your account:</strong> Profile → Delete Account (removes all your data).</li>
<li><strong>Report or block someone:</strong> on any feed post, tap the “···” menu → Report or Block.</li>
<li><strong>Notifications:</strong> turn on/off in iOS Settings → Inhale With Me → Notifications.</li>
<li><strong>Who can see a session:</strong> choose public, friends, or private when you log it.</li>
</ul>

<h2>Legal</h2>
<p><a href="/privacy">Privacy Policy</a> · <a href="/terms">Terms of Use</a> · <a href="/impressum">Impressum</a></p>`

// ---------------------------------------------------------------- Deutsch

const privacyDE = `
<h1>Datenschutzerklärung</h1>
<p class="date">Stand: 2026</p>

<h2>1. Verantwortlicher</h2>
<p>Verantwortlich für die Datenverarbeitung in dieser App:<br>Noah Feldt · Niedertorstr. 32, 32312 Lübbecke · noah@feldt.systems</p>

<h2>2. Überblick</h2>
<p>Inhale With Me (die „App“) ermöglicht das Protokollieren von Rauch-Sessions und das Teilen mit Freunden. Nachfolgend erklären wir, welche personenbezogenen Daten wir zu welchen Zwecken verarbeiten.</p>

<h2>3. Welche Daten wir verarbeiten</h2>
<ul>
<li><strong>Kontodaten:</strong> E-Mail-Adresse, Benutzername, optional Anzeigename, Bio, Avatar-URL.</li>
<li><strong>Aktivitätsdaten:</strong> protokollierte Sessions (Typ, Menge, Zeitpunkt, optional Notiz/Stimmung/Ort, Sichtbarkeit).</li>
<li><strong>Soziale Daten:</strong> Freundschaften, Anfragen, Reaktionen.</li>
<li><strong>Geräte-Token:</strong> ein Apple-Push-Token, nur wenn du Benachrichtigungen aktivierst.</li>
<li><strong>Standort (optional):</strong> nur wenn du ihn aktiv teilst, werden deine Koordinaten an den gewählten Freund gesendet und kurz (ca. 1 Stunde) gespeichert, danach verworfen. Keine Hintergrund-Verfolgung.</li>
<li><strong>Technische Daten:</strong> IP-Adresse und Server-Logs beim Zugriff auf unsere API.</li>
</ul>

<h2>4. Zwecke und Rechtsgrundlagen (Art. 6 DSGVO)</h2>
<ul>
<li>Bereitstellung der App und deines Kontos: Art. 6 Abs. 1 lit. b DSGVO (Vertrag).</li>
<li>Push-Benachrichtigungen: Art. 6 Abs. 1 lit. a DSGVO (Einwilligung); jederzeit in den iOS-Einstellungen widerrufbar.</li>
<li>Standort-Teilen: Art. 6 Abs. 1 lit. a DSGVO (Einwilligung), nur wenn du aktiv teilst.</li>
<li>Betrieb, Sicherheit, Missbrauchsvermeidung, Logs: Art. 6 Abs. 1 lit. f DSGVO (berechtigtes Interesse).</li>
</ul>

<h2>5. Empfänger</h2>
<p>Hosting: mittwald (Server in Deutschland/EU). Push-Versand: Apple Push Notification service (Apple Inc.). Ein Verkauf deiner Daten oder eine Weitergabe zu Werbezwecken findet nicht statt.</p>

<h2>6. Sichtbarkeit sozialer Inhalte</h2>
<p>Deine Sessions sind nur für dich und – je nach von dir gewählter Sichtbarkeit (öffentlich / Freunde / privat) – für deine Freunde sichtbar.</p>

<h2>7. Speicherdauer &amp; Löschung</h2>
<p>Wir speichern deine Daten, bis du sie löschst. Du kannst dein Konto und alle zugehörigen Daten jederzeit in der App unter Profil → Konto löschen entfernen.</p>

<h2>8. Deine Rechte</h2>
<p>Du hast das Recht auf Auskunft (Art. 15), Berichtigung (Art. 16), Löschung (Art. 17), Einschränkung (Art. 18), Datenübertragbarkeit (Art. 20) und Widerspruch (Art. 21) sowie das Recht, eine erteilte Einwilligung jederzeit zu widerrufen. Kontakt: noah@feldt.systems. Zudem besteht ein Beschwerderecht bei einer Datenschutz-Aufsichtsbehörde.</p>

<h2>9. Kinder</h2>
<p>Die App richtet sich an Erwachsene (17+) und nicht an Kinder.</p>

<h2>10. Änderungen</h2>
<p>Wir können diese Erklärung anpassen; es gilt jeweils die aktuelle Fassung.</p>

<p class="muted">Bitte vor Veröffentlichung rechtlich prüfen. <a href="/impressum">Impressum</a></p>`

const termsDE = `
<h1>Nutzungsbedingungen (EULA)</h1>
<p class="date">Stand: 2026</p>
<p>Mit dem Erstellen eines Kontos oder der Nutzung von Inhale With Me (die „App“) stimmst du diesen Bedingungen zu. Wenn du nicht zustimmst, nutze die App bitte nicht.</p>

<h2>1. Voraussetzungen</h2>
<p>Du musst mindestens 18 Jahre alt sein (bzw. volljährig nach dem Recht deines Wohnorts). Die App ist ein persönliches Tracking- und Social-Tool; sie verkauft, bewirbt oder fördert den Konsum von Produkten nicht.</p>

<h2>2. Dein Konto</h2>
<p>Du bist für deine Zugangsdaten und für die Inhalte verantwortlich, die du protokollierst und teilst.</p>

<h2>3. Zulässige Nutzung</h2>
<p>Missbrauche die App nicht, belästige keine anderen Nutzer und poste keine rechtswidrigen oder anstößigen Inhalte. Es gilt eine Null-Toleranz-Politik gegenüber anstößigen Inhalten und missbräuchlichem Verhalten. Du kannst Inhalte melden und Nutzer blockieren; Meldungen werden geprüft und Inhalte oder Konten ggf. entfernt.</p>

<h2>4. Nutzerinhalte</h2>
<p>Du behältst das Eigentum an deinen Inhalten und räumst uns ein begrenztes Recht ein, sie zu speichern und dir sowie – gemäß deinen Sichtbarkeitseinstellungen – deinen Freunden anzuzeigen, ausschließlich zum Betrieb der App.</p>

<h2>5. Konto-Löschung</h2>
<p>Du kannst dein Konto jederzeit unter Profil → Konto löschen entfernen; deine Daten werden dabei dauerhaft gelöscht.</p>

<h2>6. Haftungsausschluss</h2>
<p>Die App wird „wie besehen“ ohne Gewährleistung bereitgestellt und stellt keine medizinische Beratung dar. Soweit gesetzlich zulässig, haften wir nicht für Schäden aus der Nutzung der App.</p>

<h2>7. Änderungen &amp; Kontakt</h2>
<p>Wir können diese Bedingungen anpassen; die weitere Nutzung gilt als Zustimmung. Kontakt: noah@feldt.systems</p>`

const supportDE = `
<h1>Support</h1>
<p>Brauchst du Hilfe mit Inhale With Me? Wir helfen gern.</p>

<h2>Kontakt</h2>
<p>E-Mail <a href="mailto:noah@feldt.systems">noah@feldt.systems</a>. Antwort in der Regel innerhalb weniger Tage.</p>

<h2>Häufige Fragen</h2>
<ul>
<li><strong>Konto löschen:</strong> Profil → Konto löschen (entfernt alle Daten).</li>
<li><strong>Melden oder blockieren:</strong> bei jedem Feed-Beitrag auf das „···“-Menü → Melden oder Blockieren.</li>
<li><strong>Benachrichtigungen:</strong> ein-/ausschalten in iOS-Einstellungen → Inhale With Me → Mitteilungen.</li>
<li><strong>Wer eine Session sieht:</strong> beim Loggen öffentlich, Freunde oder privat wählen.</li>
</ul>

<h2>Rechtliches</h2>
<p><a href="/privacy/de">Datenschutzerklärung</a> · <a href="/terms/de">Nutzungsbedingungen</a> · <a href="/impressum">Impressum</a></p>`
