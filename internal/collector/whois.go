package collector

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/likexian/whois"
	"github.com/typicalfo/netgaze/internal/model"
)

func collectWhois(ctx context.Context, target string, report *model.Report) error {
	// Create context with 10-second timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Run WHOIS in goroutine to respect context
	resultChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	go func() {
		result, err := whois.Whois(target)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	// Wait for completion or timeout
	select {
	case result := <-resultChan:
		report.WhoisRaw = result
		parseWhoisData(result, report)
	case err := <-errorChan:
		report.Errors["whois"] = fmt.Sprintf("WHOIS failed: %v", err)
		// Don't return error for WHOIS - it's optional
		return nil
	case <-ctx.Done():
		report.Errors["whois"] = "WHOIS timeout"
		// Don't return error for WHOIS - it's optional
		return nil
	}

	return nil
}

func parseWhoisData(data string, report *model.Report) {
	// Convert to lowercase for case-insensitive matching
	lowerData := strings.ToLower(data)

	// Parse common WHOIS fields using regex patterns
	report.Whois.Domain = extractField(data, lowerData, []string{
		`domain name:\s*(.+)`,
		`domain:\s*(.+)`,
	})

	report.Whois.Registrar = extractField(data, lowerData, []string{
		`registrar:\s*(.+)`,
		`registrar name:\s*(.+)`,
		`sponsoring registrar:\s*(.+)`,
	})

	report.Whois.Created = extractField(data, lowerData, []string{
		`creation date:\s*(.+)`,
		`created:\s*(.+)`,
		`registered:\s*(.+)`,
		`registration time:\s*(.+)`,
	})

	report.Whois.Expires = extractField(data, lowerData, []string{
		`expiration date:\s*(.+)`,
		`expires:\s*(.+)`,
		`expiry date:\s*(.+)`,
		`paid-till:\s*(.+)`,
	})

	report.Whois.Registrant = extractField(data, lowerData, []string{
		`registrant name:\s*(.+)`,
		`registrant organization:\s*(.+)`,
		`registrant:\s*(.+)`,
	})

	// For IP addresses, parse network information
	report.Whois.NetRange = extractField(data, lowerData, []string{
		`inetnum:\s*(.+)`,
		`netrange:\s*(.+)`,
		`cidr:\s*(.+)`,
		`route:\s*(.+)`,
	})

	report.Whois.NetName = extractField(data, lowerData, []string{
		`netname:\s*(.+)`,
		`network name:\s*(.+)`,
	})

	report.Whois.OrgName = extractField(data, lowerData, []string{
		`organization:\s*(.+)`,
		`org:\s*(.+)`,
		`descr:\s*(.+)`,
	})

	report.Whois.Country = extractField(data, lowerData, []string{
		`country:\s*(.+)`,
		`registrant country:\s*(.+)`,
	})

	// Extract abuse emails
	report.Whois.AbuseEmails = extractEmails(data)
}

func extractField(data, lowerData string, patterns []string) string {
	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern) // Case-insensitive
		matches := re.FindStringSubmatch(lowerData)
		if len(matches) > 1 {
			// Try to get original case from data
			originalMatch := regexp.MustCompile(`(?i)` + pattern).FindStringSubmatch(data)
			if len(originalMatch) > 1 {
				return strings.TrimSpace(originalMatch[1])
			}
			return strings.TrimSpace(matches[1])
		}
	}
	return ""
}

func extractEmails(data string) []string {
	// Common email patterns in WHOIS
	emailPatterns := []string{
		`abuse.*?([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`,
		`admin.*?([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`,
		`technical.*?([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`,
		`([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`,
	}

	var emails []string
	seen := make(map[string]bool)

	for _, pattern := range emailPatterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		matches := re.FindAllStringSubmatch(data, -1)

		for _, match := range matches {
			if len(match) > 1 {
				email := strings.ToLower(strings.TrimSpace(match[1]))
				if !seen[email] && strings.Contains(email, "@") {
					emails = append(emails, email)
					seen[email] = true
				}
			}
		}
	}

	return emails
}
