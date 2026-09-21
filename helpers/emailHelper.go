// Package helpers contains small reusable pieces of logic
// shared across handlers (token generation, email sending, etc).
package helpers

import (
	"fmt"
	"log"

	"peter-go-auth-template/initializers"

	"github.com/resend/resend-go/v4"
)

// SendPasswordResetEmail sends a password reset link to the given address
// using the Resend API.
//
// Until a custom domain is verified in Resend, RESEND_FROM_EMAIL must be
// "onboarding@resend.dev", and the recipient must be the same email used
// to sign up to Resend. Otherwise the API returns 403.
//
// SECURITY: never log the raw token in production — it would leak into
// log aggregation systems and become a credential.
func SendPasswordResetEmail(to string, rawToken string) error {
	client := resend.NewClient(initializers.ResendAPIKey)

	// Where the click lands. Right now this points at localhost, which
	// won't render a page — but the token is in the URL, so a developer
	// testing with Postman can copy it out of the email body directly.
	resetLink := fmt.Sprintf(
		"%s/reset-password?token=%s",
		initializers.FrontendURL,
		rawToken,
	)

	// Inline-styled HTML. Email clients strip <style> tags and external
	// CSS, so styles have to live on the elements themselves.
	// Minimal template — no tracking pixels, no external assets.
	htmlBody := fmt.Sprintf(`
		<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 520px; margin: 0 auto; padding: 24px;">
			<h2 style="color: #111;">Reset your password</h2>
			<p style="color: #444; line-height: 1.6;">
				We received a request to reset the password for your account.
				Click the button below to choose a new one. This link expires in 15 minutes.
			</p>
			<p style="margin: 32px 0;">
				<a href="%s"
				   style="background: #111; color: #fff; padding: 12px 24px;
				          border-radius: 6px; text-decoration: none; font-weight: 600;">
					Reset password
				</a>
			</p>
			<p style="color: #888; font-size: 13px; line-height: 1.6;">
				If you didn't request this, you can safely ignore this email.
				Your password won't change until you click the link above.
			</p>
			<hr style="border: none; border-top: 1px solid #eee; margin: 24px 0;" />
			<p style="color: #aaa; font-size: 12px;">
				If the button doesn't work, copy and paste this URL into your browser:<br />
				<span style="color: #666; word-break: break-all;">%s</span>
			</p>
		</div>
	`, resetLink, resetLink)

	params := &resend.SendEmailRequest{
		From:    initializers.ResendFromEmail,
		To:      []string{to},
		Subject: "Reset your password",
		Html:    htmlBody,
	}

	// Emails.Send blocks until Resend responds. For higher throughput
	// you'd push this into a background queue, but sync is fine here.
	sent, err := client.Emails.Send(params)
	if err != nil {
		log.Printf("[RESEND ERROR] to=%s err=%v", to, err)
		return err
	}

	// Log the message ID so you can look up the delivery in the Resend
	// dashboard (Events page) if a user says the email never arrived.
	log.Printf("[RESEND OK] to=%s id=%s", to, sent.Id)
	return nil
}
