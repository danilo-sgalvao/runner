package com.hubsaude.assinador.presentation.cli;

import com.hubsaude.assinador.domain.model.SignatureResult;
import com.hubsaude.assinador.infrastructure.json.JsonMapper;

public class CliPresenter {

    public void present(SignatureResult result) {
        String json = JsonMapper.toJson(result);
        if (result.isValid()) {
            System.out.println(json);
        } else {
            System.err.println(json);
            System.exit(1);
        }
    }

    public void presentError(String message) {
        System.err.println(message);
    }
}
