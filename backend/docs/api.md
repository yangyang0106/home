# API Draft

Base URL: `/api/v1`

## Health

- `GET /health`

## Meta

- `GET /meta`
- 返回指标定义、房屋类型映射、装修映射、加分项和风险项

## Dashboard

- `GET /households/{householdId}/dashboard`
- 返回：
  - 当前家庭的双人权重
  - 全部房源
  - 后端计算后的排序结果
  - summary 总览数据

## Weights

- `GET /households/{householdId}/weights`
- `PUT /households/{householdId}/weights`

Request:

```json
{
  "profiles": [
    {
      "role": "me",
      "label": "我的偏好",
      "weights": {
        "totalPrice": 25,
        "commuteTime": 22
      }
    }
  ]
}
```

## Houses

- `GET /households/{householdId}/houses`
- `POST /households/{householdId}/houses`
- `GET /households/{householdId}/houses/{houseId}`
- `PUT /households/{householdId}/houses/{houseId}`
- `DELETE /households/{householdId}/houses/{houseId}`

House payload:

```json
{
  "id": "house-1",
  "householdId": "demo-family",
  "communityName": "春申景城",
  "listingName": "3号楼 1202",
  "viewDate": "2026-03-17",
  "totalPrice": 698,
  "unitPrice": 76500,
  "area": 91.2,
  "houseAge": 11,
  "floor": "18F 中层",
  "orientation": "南北",
  "houseType": "商品房",
  "renovation": "精装",
  "commuteTime": 42,
  "metroTime": 10,
  "monthlyFee": 580,
  "livingConvenience": 8,
  "efficiencyRate": 79,
  "lightScore": 8,
  "noiseScore": 7,
  "layoutScore": 8,
  "propertyScore": 8,
  "communityScore": 8,
  "comfortScore": 8,
  "parkingScore": 7,
  "bonusSelections": ["charger", "parkingSpot"],
  "riskSelections": ["secondaryRoad"],
  "notes": "厨房略小，但整体顺眼。"
}
```
