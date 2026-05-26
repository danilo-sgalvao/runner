package com.hubsaude.assinador.infrastructure.json;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

public final class JsonMapper {

    private static final ObjectMapper MAPPER = new ObjectMapper();

    private JsonMapper() {}

    public static String toJson(Object value) {
        try {
            return MAPPER.writeValueAsString(value);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Falha ao serializar resposta em JSON", e);
        }
    }
}
