package com.hubsaude.assinador.domain;

public class SignRequest {

    private String content;
    private String token;

    public SignRequest() {}

    public String getContent() {
        return content;
    }

    public void setContent(String content) {
        this.content = content;
    }

    public String getToken() {
        return token;
    }

    public void setToken(String token) {
        this.token = token;
    }
}
