package logger

// Base field builders
func BaseFields() Fields {
    return Fields{
        "service": "mMAD-engines",
        "version": "1.0.0",
    }
}

func ComponentFields(component string) Fields {
    return Fields{
        "component": component,
    }
}

func ProofFields(proofType, circuitName, proofID string) Fields {
    fields := ComponentFields("zkproof")
    fields["proof_type"] = proofType
    fields["circuit_name"] = circuitName
    fields["proof_id"] = proofID
    return fields
}

func ReserveFields(amount, threshold, bankAccount string) Fields {
    fields := ComponentFields("reserve")
    fields["reserve_amount"] = amount
    fields["reserve_threshold"] = threshold
    fields["bank_account"] = bankAccount
    return fields
}

func ComplianceFields(userID, checkType, status string) Fields {
    fields := ComponentFields("compliance")
    fields["user_id"] = userID
    fields["check_type"] = checkType
    fields["status"] = status
    return fields
}

func BlockchainFields(txHash, contractAddress, blockNumber string) Fields {
    fields := ComponentFields("blockchain")
    fields["tx_hash"] = txHash
    fields["contract_address"] = contractAddress
    fields["block_number"] = blockNumber
    return fields
}

func PerformanceFields(duration, operation string) Fields {
    return Fields{
        "duration_ms": duration,
        "operation": operation,
        "type": "performance",
    }
}

func ErrorFields(errorCode, errorType string, err error) Fields {
    fields := Fields{
        "error_code": errorCode,
        "error_type": errorType,
    }
    if err != nil {
        fields["error_message"] = err.Error()
    }
    return fields
}

func MergeFields(fieldSets ...Fields) Fields {
    result := make(Fields)
    for _, fields := range fieldSets {
        for k, v := range fields {
            result[k] = v
        }
    }
    return result
}