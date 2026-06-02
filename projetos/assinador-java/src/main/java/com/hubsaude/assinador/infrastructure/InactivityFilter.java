package com.hubsaude.assinador.infrastructure;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;

/** Atualiza o timestamp de última requisição a cada chamada HTTP recebida. */
@Component
public class InactivityFilter extends OncePerRequestFilter {

    private final RequestTimestamp requestTimestamp;

    public InactivityFilter(RequestTimestamp requestTimestamp) {
        this.requestTimestamp = requestTimestamp;
    }

    @Override
    protected void doFilterInternal(HttpServletRequest request,
                                    HttpServletResponse response,
                                    FilterChain filterChain) throws ServletException, IOException {
        requestTimestamp.touch();
        filterChain.doFilter(request, response);
    }
}
