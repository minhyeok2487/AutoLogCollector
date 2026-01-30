# 화면별 상세 가이드

[← 목차로 돌아가기](./USER_GUIDE.md)

---

## Execution 화면

메인 화면으로, 접속 정보 설정부터 실행까지 모든 작업을 수행합니다.

![Execution 화면](../01.png)

### Connection Settings

| 항목 | 설명 | 기본값 |
|------|------|--------|
| Username | SSH 접속 계정 | - |
| Password | SSH 접속 비밀번호 | - |
| Timeout | 접속 및 명령 응답 대기 시간 (초) | 10 |
| Disable Paging | `terminal length 0` 자동 전송 | 활성화 |
| Auto Export Excel | 완료 시 자동 Excel 생성 | 비활성화 |
| Enable Mode | 특권 모드 진입 | 비활성화 |
| Enable Password | Enable Mode 비밀번호 (Enable Mode 활성화 시 표시) | - |

### Target Servers

서버 목록을 관리하는 영역입니다.

- **Import CSV**: `ip,hostname` 형식의 CSV 파일에서 서버 목록을 일괄 불러옵니다.
- **Export CSV**: 현재 서버 목록을 CSV 파일로 내보냅니다.
- **+ Add**: 테이블에 새 행을 추가하여 IP와 Hostname을 직접 입력합니다.
- **삭제**: 각 행의 삭제 버튼으로 개별 서버를 제거합니다.
- **서버별 인증(🔒)**: 잠금 아이콘을 클릭하면 해당 서버에만 적용되는 별도 인증 정보를 설정할 수 있습니다. 자세한 내용은 [고급 기능 - 서버별 개별 인증](./03-advanced.md#서버별-개별-인증-per-server-credentials)을 참조하세요.

### Commands

- 텍스트 영역에 명령어를 한 줄씩 입력합니다.
- **Import TXT**: 텍스트 파일에서 명령어를 불러옵니다.
- **Export TXT**: 현재 명령어를 텍스트 파일로 저장합니다.
- 화면 하단에 입력된 명령어 수가 표시됩니다.

### 실행

- **Run Execution**: 설정된 서버 목록에 SSH 접속하여 명령어를 실행합니다.
- **Stop**: 실행 중 중단합니다. 이미 진행 중인 서버 작업은 완료된 후 중단됩니다.
- 진행률 바와 완료 서버 수가 실시간으로 표시됩니다.

---

## Results 화면

![Results 화면](../02.png)

실행이 완료되면 서버별 결과를 확인할 수 있습니다.

| 열 | 설명 |
|----|------|
| Hostname | 서버 호스트명 |
| IP | 서버 IP 주소 |
| Status | 성공(Success) 또는 실패(Failed) |
| Duration | 명령 실행 소요 시간 |
| Action | View Log 버튼 |

- **View Log**: 해당 서버의 전체 로그를 모달 창에서 확인합니다.
- **Export Excel**: 명령어별로 시트가 구분된 Excel 파일을 생성합니다. 각 열은 서버, 각 행은 해당 명령의 출력입니다.
- **Open Logs Folder**: Windows 파일 탐색기에서 로그 폴더를 엽니다.

상단 요약 바에서 성공/실패/전체 서버 수를 한눈에 확인할 수 있습니다.

---

## Live Logs 화면

실행 중 모든 서버의 SSH 세션 출력을 실시간으로 확인합니다.

- **Combined 탭**: 모든 서버의 로그가 타임스탬프와 서버명과 함께 통합 표시됩니다.
- **서버별 탭**: 실행이 시작되면 각 서버 이름으로 탭이 자동 생성됩니다. 탭을 클릭하면 해당 서버의 로그만 필터링됩니다.
- **Auto-scroll**: 활성화하면 새 로그 수신 시 자동으로 하단으로 스크롤합니다.
- **Clear**: 화면의 로그를 초기화합니다 (파일에 저장된 로그는 유지).

---

## Schedule 화면

등록된 스케줄 목록을 관리합니다.

| 열 | 설명 |
|----|------|
| Name | 스케줄 이름 |
| Type | Daily / Weekly / Monthly |
| Time | 실행 시각 |
| Next Run | 다음 실행 예정 시각 |
| Last Run | 마지막 실행 시각 |
| Status | Enabled / Disabled |
| Actions | 편집, 삭제, 활성화/비활성화 |

자세한 내용은 [스케줄링 완전 가이드](./04-scheduling.md)를 참조하세요.

---

[← 빠른 시작 가이드](./01-quick-start.md) | [다음: 고급 기능 →](./03-advanced.md)
