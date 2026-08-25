package api

import "net/http"

// handleTerms and handlePrivacy serve public legal pages. They are linked from
// the app (Profile) and used for the App Store privacy-policy URL. Replace the
// placeholder text with reviewed legal copy before public release.

func (a *API) handleTerms(w http.ResponseWriter, r *http.Request) {
	writeHTML(w, termsHTML)
}

func (a *API) handlePrivacy(w http.ResponseWriter, r *http.Request) {
	writeHTML(w, privacyHTML)
}

func writeHTML(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(body))
}

const termsHTML = `<!doctype html>
<html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Inhale With Me — Terms of Use</title>
<style>body{font:16px/1.6 -apple-system,system-ui,sans-serif;max-width:720px;margin:2rem auto;padding:0 1rem;color:#111}h1,h2{line-height:1.25}</style>
</head><body>
<h1>Terms of Use (EULA)</h1>
<p><em>Last updated: 2026.</em></p>
<p>By creating an account or using Inhale With Me (the “App”), you agree to these Terms. If you do not agree, do not use the App.</p>
<h2>1. Eligibility</h2>
<p>You must be at least 18 years old (or the age of legal majority where you live) to use the App. The App is a personal tracking and social tool; it does not sell, promote, or encourage the consumption of any product.</p>
<h2>2. Your account</h2>
<p>You are responsible for your credentials and for the content you log and share. Keep your password secure.</p>
<h2>3. Acceptable use</h2>
<p>You agree not to misuse the App, harass other users, post unlawful or objectionable content, or attempt to disrupt the service. There is zero tolerance for objectionable content or abusive behavior. You can report content and block users in the App; reports are reviewed and offending content or accounts may be removed.</p>
<h2>4. User content</h2>
<p>You retain ownership of what you create. You grant us a limited license to store and display it to you and, per your visibility settings, to your friends, solely to operate the App.</p>
<h2>5. Account deletion</h2>
<p>You can delete your account at any time from Profile → Delete Account, which permanently removes your data.</p>
<h2>6. Disclaimer</h2>
<p>The App is provided “as is,” without warranties. It is not medical advice. To the extent permitted by law, we are not liable for damages arising from your use of the App.</p>
<h2>7. Changes</h2>
<p>We may update these Terms; continued use after changes constitutes acceptance.</p>
<h2>8. Contact</h2>
<p>Questions: noah@feldt.systems</p>
</body></html>`

const privacyHTML = `<!doctype html>
<html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Inhale With Me — Privacy Policy</title>
<style>body{font:16px/1.6 -apple-system,system-ui,sans-serif;max-width:720px;margin:2rem auto;padding:0 1rem;color:#111}h1,h2{line-height:1.25}</style>
</head><body>
<h1>Privacy Policy</h1>
<p><em>Last updated: 2026.</em></p>
<p>This policy explains what Inhale With Me (the “App”) collects and how it is used.</p>
<h2>Data we collect</h2>
<ul>
<li><strong>Account:</strong> email, username, optional display name, bio, avatar URL.</li>
<li><strong>Activity you log:</strong> session type, quantity, time, optional note/mood/location, and visibility.</li>
<li><strong>Social:</strong> friends, friend requests, and reactions.</li>
<li><strong>Device token:</strong> an Apple push token, only if you enable notifications, used to deliver pushes.</li>
</ul>
<h2>How we use it</h2>
<p>To provide the App: authenticate you, show your stats, share activity with friends per your visibility settings, and send notifications you opt into. We do <strong>not</strong> sell your personal data.</p>
<h2>Sharing</h2>
<p>Your activity is visible only to you and, per each entry’s visibility, to your friends. Push notifications are delivered via Apple (APNs). Data is hosted on our provider (mittwald) in the EU.</p>
<h2>Retention &amp; deletion</h2>
<p>We keep your data until you delete it. You can delete your account and all associated data at any time from Profile → Delete Account.</p>
<h2>Your rights</h2>
<p>You may access, correct, or delete your data. For requests, contact us below.</p>
<h2>Children</h2>
<p>The App is for adults (18+) and is not directed to children.</p>
<h2>Contact</h2>
<p>noah@feldt.systems</p>
</body></html>`
