package com.hubsaude.assinador.presentation.http.dto;

public class ValidateHttpRequest {

    private String content;
    private String signature;

    public String getContent() { return content; }
    public void setContent(String content) { this.content = content; }

    public String getSignature() { return signature; }
    public void setSignature(String signature) { this.signature = signature; }
}
