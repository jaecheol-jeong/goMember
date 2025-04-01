# Member Management API

이 프로젝트는 회원 관리 시스템을 위한 API를 제공합니다. 아래는 주요 엔드포인트와 그 기능에 대한 요약입니다.

## API 요약

### 1. 회원 생성
- **메서드**: POST
- **URL**: `/api/members`
- **설명**: 새로운 회원을 생성합니다.
- **요청 예시**:
  ```json
  {
    "name": "홍길동",
    "email": "hong@example.com"
  }
  ```
- **응답 예시**:
  ```json
  {
    "id": 1,
    "name": "홍길동",
    "email": "hong@example.com"
  }
  ```

### 2. 회원 조회
- **메서드**: GET
- **URL**: `/api/members/{id}`
- **설명**: 특정 회원의 정보를 조회합니다.
- **응답 예시**:
  ```json
  {
    "id": 1,
    "name": "홍길동",
    "email": "hong@example.com"
  }
  ```

### 3. 회원 목록 조회
- **메서드**: GET
- **URL**: `/api/members`
- **설명**: 모든 회원의 목록을 조회합니다.
- **응답 예시**:
  ```json
  [
    {
      "id": 1,
      "name": "홍길동",
      "email": "hong@example.com"
    },
    {
      "id": 2,
      "name": "김철수",
      "email": "kim@example.com"
    }
  ]
  ```

### 4. 회원 정보 수정
- **메서드**: PUT
- **URL**: `/api/members/{id}`
- **설명**: 특정 회원의 정보를 수정합니다.
- **요청 예시**:
  ```json
  {
    "name": "홍길동 수정",
    "email": "hong_updated@example.com"
  }
  ```
- **응답 예시**:
  ```json
  {
    "id": 1,
    "name": "홍길동 수정",
    "email": "hong_updated@example.com"
  }
  ```

### 5. 회원 삭제
- **메서드**: DELETE
- **URL**: `/api/members/{id}`
- **설명**: 특정 회원을 삭제합니다.
- **응답 예시**:
  ```json
  {
    "message": "회원이 삭제되었습니다."
  }
  ```

## 사용 방법
1. API 서버를 실행합니다.
2. 위의 엔드포인트를 사용하여 회원 관리 작업을 수행합니다.

