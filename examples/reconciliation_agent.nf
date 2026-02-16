\ ==============================================================================
\ MODULE: Automated Financial Reconciliation
\ DESCRIPTION: Audits external payment gateways against internal ledgers.
\ ==============================================================================

\ ------------------------------------------------------------------------------
\ HELPER WORD: Pure Logic Comparison (Runs in the Root Sandbox)
\ ------------------------------------------------------------------------------
: FIND-MISSING-TXNS { external-list internal-list }
    \ No capabilities needed here. The agent is completely demilitarized.
    external-list internal-list CROSS-REFERENCE INTO diff-result
    diff-result EXTRACT-MISSING INTO missing-items
    
    \ YIELD explicitly pushes the named variable back to the ephemeral 
    \ transit layer for the caller to capture.
    missing-items YIELD 
;

\ ------------------------------------------------------------------------------
\ MAIN ORCHESTRATOR: The Agentic Loop
\ The Go Host VM injects 4 strictly scoped capability tokens onto the stack.
\ ------------------------------------------------------------------------------
: RECONCILE-PAYMENTS { gateway-cap sql-cap fs-cap slack-cap }

    \ --- 1. EXTERNAL EXTRACTION (Payment Gateway) ---
    ADDRESS HTTP-ENV gateway-cap
    
    WITH-CLIENT
        SET-URL "https://api.payment-gateway.external/v1/settlements/today"
        SET-METHOD "GET"
        SEND-REQUEST 
    INTO gateway-response
    
    gateway-response CHECK-STATUS 200 != IF
        "CRITICAL: Payment Gateway API unreachable." THROW
    THEN
    
    gateway-response PARSE-JSON INTO external-txns
    
    \ --- 2. INTERNAL EXTRACTION (Secure Database) ---
    \ Notice: The HTTP environment has implicitly closed. We open SQL.
    ADDRESS SQL-ENV sql-cap
    
    PREPARE-QUERY "SELECT txn_id, amount FROM ledger WHERE date = CURRENT_DATE"
    EXECUTE-QUERY INTO db-cursor
    
    db-cursor FETCH-ALL INTO internal-txns
    
    \ --- 3. TRANSFORMATION (Data Comparison) ---
    \ We call our helper word and ground the YIELDed result into a new local state
    external-txns internal-txns FIND-MISSING-TXNS INTO missing-records
    
    missing-records IS-EMPTY IF
        "Reconciliation successful. Ledgers match perfectly." PRINT
        EXIT
    THEN

    \ --- 4. CONDITIONAL ACTUATION & ALERTING ---
    \ If we reach here, money is missing. 
    
    \ 4A. Write the forensic report
    ADDRESS FS-ENV fs-cap
    
    \ The fs-cap restricts writes strictly to /var/finance/audits/
    OPEN-FILE "/var/finance/audits/discrepancy_report.json" WITH-MODE "WRITE" INTO audit-log
    audit-log missing-records WRITE-JSON
    audit-log CLOSE-FILE
    
    \ 4B. Alert the Finance Team
    \ CRITICAL SECURITY BOUNDARY: We re-open HTTP, but MUST use 'slack-cap'
    ADDRESS HTTP-ENV slack-cap
    
    missing-records COUNT-ITEMS INTO missing-count
    "URGENT: %s missing transactions detected!" missing-count FORMAT-STRING INTO alert-msg
    
    WITH-CLIENT
        SET-URL "https://hooks.slack.com/services/FINANCE/ALERTS"
        SET-METHOD "POST"
        alert-msg ADD-JSON-PAYLOAD "text"
        SEND-REQUEST
    INTO slack-result
    
    "Discrepancies found and reported. Workflow halted." PRINT
;
