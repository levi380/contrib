// func Now(physical bool) uint64
TEXT ·Now(SB), $0
    MOVB    physical+0(FP), R0
    CMP $0, R0
    BEQ 3(PC)
    
    MRS CNTPCT_EL0, R0
    B   2(PC)

    MRS CNTVCT_EL0, R0
    MOVD    R0, ret+8(FP)
    RET
