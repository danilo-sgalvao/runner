package com.hubsaude.assinador.presentation.http.dto;

public class SignHttpRequest {

    private String content;
    private String token;

    public String getContent() { return content; }
    public void setContent(String content) { this.content = content; }

    public String getToken() { return token; }
    public void setToken(String token) { this.token = token; }
}
