# Yellow: share float precision

Complementary share updates use three decimal places in the UI. Very small float residuals can still fail the `0.001` server tolerance if users type many digits; clamp/round on blur if that appears in practice.
