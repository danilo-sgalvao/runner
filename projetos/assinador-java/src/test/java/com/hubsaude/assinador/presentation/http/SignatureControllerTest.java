package com.hubsaude.assinador.presentation.http;

import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@SpringBootTest
@AutoConfigureMockMvc
class SignatureControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @Test
    void postSign_contentValido_retorna200ComAssinatura() throws Exception {
        mockMvc.perform(post("/sign")
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"content\":\"documento\"}"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.valid").value(true))
                .andExpect(jsonPath("$.signature").value("MOCKED_SIGNATURE_BASE64_=="))
                .andExpect(jsonPath("$.message").value("Assinatura criada com sucesso"));
    }

    @Test
    void postSign_contentVazio_retorna400() throws Exception {
        mockMvc.perform(post("/sign")
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"content\":\"\"}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.valid").value(false))
                .andExpect(jsonPath("$.signature").isEmpty())
                .andExpect(jsonPath("$.message").value("Parâmetro 'content' inválido ou ausente"));
    }

    @Test
    void postSign_contentAusente_retorna400() throws Exception {
        mockMvc.perform(post("/sign")
                .contentType(MediaType.APPLICATION_JSON)
                .content("{}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.valid").value(false));
    }

    @Test
    void postValidate_assinaturaCorreta_retorna200Valid() throws Exception {
        mockMvc.perform(post("/validate")
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"content\":\"doc\",\"signature\":\"MOCKED_SIGNATURE_BASE64_==\"}"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.valid").value(true))
                .andExpect(jsonPath("$.message").value("Assinatura é válida"));
    }

    @Test
    void postValidate_assinaturaErrada_retorna200Invalid() throws Exception {
        mockMvc.perform(post("/validate")
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"content\":\"doc\",\"signature\":\"assinatura-errada\"}"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.valid").value(false))
                .andExpect(jsonPath("$.message").value("Assinatura é inválida"));
    }

    @Test
    void postValidate_contentAusente_retorna400() throws Exception {
        mockMvc.perform(post("/validate")
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"signature\":\"MOCKED_SIGNATURE_BASE64_==\"}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.valid").value(false))
                .andExpect(jsonPath("$.message").value("Parâmetro 'content' inválido ou ausente"));
    }

    @Test
    void postValidate_signatureAusente_retorna400() throws Exception {
        mockMvc.perform(post("/validate")
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"content\":\"doc\"}"))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.valid").value(false))
                .andExpect(jsonPath("$.message").value("Parâmetro 'signature' inválido ou ausente"));
    }
}
