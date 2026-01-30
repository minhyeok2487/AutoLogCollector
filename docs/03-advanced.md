# 고급 기능

[← 목차로 돌아가기](./USER_GUIDE.md)

---

## 병렬 실행 (Concurrent) 설정

Concurrent 값은 동시에 SSH 접속을 시도하는 서버 수를 제어합니다.

- 서버 수가 적은 경우(10대 이하): 전체 서버 수와 동일하게 설정
- 서버 수가 많은 경우(10대 이상): 5~10 정도로 설정하여 네트워크 부하 분산
- 너무 높은 값은 네트워크 병목이나 장비 부하를 유발할 수 있습니다

---

## 서버별 개별 인증 (Per-server Credentials)

환경에 따라 서버마다 다른 계정이 필요한 경우 사용합니다.

1. Target Servers 테이블에서 해당 서버의 **🔒** 아이콘 클릭
2. 팝업에서 Username, Password, Enable Password 입력
3. **Save** 클릭

설정된 서버는 전역 인증 정보 대신 개별 인증 정보를 사용합니다. 개별 인증을 제거하면 전역 인증으로 돌아갑니다.

인증 정보는 `config/servers.json`에 AES-256-GCM으로 암호화되어 저장됩니다.

---

## Enable Mode 사용법

`show running-config` 등 특권 모드가 필요한 명령을 실행할 때 사용합니다.

1. Connection Settings에서 **Enable Mode** 체크박스 활성화
2. **Enable Password** 입력 (로그인 비밀번호와 동일한 경우 비워둘 수 있음)
3. 실행 시 자동으로 `enable` 명령을 전송하고 비밀번호를 입력합니다

---

## Disable Paging 동작 원리

Cisco 장비는 기본적으로 출력이 길면 `--More--` 프롬프트로 페이징합니다. Disable Paging을 활성화하면:

1. 세션 시작 시 `terminal length 0` 명령을 자동 전송
2. 이후 모든 명령의 출력이 끊기지 않고 전체 표시
3. 만약 `--More--` 프롬프트가 감지되면 자동으로 스페이스를 전송하여 계속 진행

---

## Auto Export Excel

활성화하면 모든 서버의 실행이 완료된 직후 자동으로 `results.xlsx` 파일이 로그 폴더에 생성됩니다. 수동으로 Results 화면에서 **Export Excel** 버튼을 클릭할 필요가 없습니다.

---

[← 화면별 상세 가이드](./02-screens.md) | [다음: 스케줄링 완전 가이드 →](./04-scheduling.md)
