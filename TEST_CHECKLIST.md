# Test Checklist - Parallel Execution Feature

## GUI Tests

- [ ] Concurrent 입력 필드가 표시되는지 확인
- [ ] 기본값이 5로 설정되어 있는지 확인
- [ ] 1~50 범위 제한이 동작하는지 확인
- [ ] Concurrent 값 1로 설정 시 순차 실행 확인
- [ ] Concurrent 값 5로 설정 시 5개 병렬 실행 확인
- [ ] Concurrent 값 10으로 설정 시 10개 병렬 실행 확인
- [ ] 실행 중 Concurrent 입력 필드가 비활성화되는지 확인
- [ ] 상태 표시에 병렬 개수가 표시되는지 확인 (예: "Running (5 parallel)...")
- [ ] Stop 버튼으로 병렬 실행 중단이 되는지 확인
- [ ] 진행률 표시가 올바르게 업데이트되는지 확인
- [ ] 결과 테이블에 모든 서버 결과가 표시되는지 확인

## CLI Tests

- [ ] `-c` 플래그 없이 실행 시 기본값 5로 동작하는지 확인
- [ ] `-c 1` 옵션으로 순차 실행 확인
- [ ] `-c 10` 옵션으로 10개 병렬 실행 확인
- [ ] 시작 시 "Concurrent connections: N" 메시지 출력 확인

## Edge Cases

- [ ] 서버 수보다 큰 Concurrent 값 설정 시 정상 동작 확인
- [ ] Concurrent 값 0 또는 음수 입력 시 기본값 1로 처리되는지 확인
- [ ] 일부 서버 실패 시 다른 서버는 계속 실행되는지 확인
- [ ] 모든 서버 실패 시 결과 요약이 올바른지 확인

## Performance

- [ ] 10개 서버, Concurrent 1 vs 10 실행 시간 비교
- [ ] 로그 파일이 각 서버별로 정상 생성되는지 확인
- [ ] 동시 실행 시 로그 파일 내용이 섞이지 않는지 확인
