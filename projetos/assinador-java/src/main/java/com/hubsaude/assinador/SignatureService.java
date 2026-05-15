package com.hubsaude.assinador;

import com.hubsaude.assinador.domain.SignRequest;
import com.hubsaude.assinador.domain.ValidateRequest;
import com.hubsaude.assinador.domain.SignatureResponse;

public interface SignatureService {
    SignatureResponse sign(SignRequest request);
    SignatureResponse validate(ValidateRequest request);
}
