package com.hubsaude.assinador.infrastructure;

import org.springframework.stereotype.Component;

import java.util.concurrent.atomic.AtomicLong;

/** Registra o instante da última requisição HTTP recebida pelo servidor. */
@Component
public class RequestTimestamp {

    private final AtomicLong lastRequest = new AtomicLong(System.currentTimeMillis());

    public void touch() {
        lastRequest.set(System.currentTimeMillis());
    }

    public long get() {
        return lastRequest.get();
    }
}
