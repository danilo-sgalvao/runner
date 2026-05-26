package com.hubsaude.assinador.presentation.http.dto;

public class SignatureHttpResponse {

    private String signature;
    private boolean valid;
    private String message;

    public SignatureHttpResponse(String signature, boolean valid, String message) {
        this.signature = signature;
        this.valid = valid;
        this.message = message;
    }

    public String getSignature() { return signature; }
    public boolean isValid() { return valid; }
    public String getMessage() { return message; }
}
